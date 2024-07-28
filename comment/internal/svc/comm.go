package svc

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/comment/internal/external"
	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/comment/internal/model"
	"github.com/ryanreadbooks/whimer/comment/internal/repo"
	"github.com/ryanreadbooks/whimer/comment/internal/repo/comm"
	"github.com/ryanreadbooks/whimer/comment/internal/repo/queue"
	"github.com/ryanreadbooks/whimer/comment/sdk"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xnet"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	notesdk "github.com/ryanreadbooks/whimer/note/sdk"

	seqer "github.com/ryanreadbooks/folium/sdk"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

const (
	seqerReplyKey = "reply-id-seqer"
)

type CommentSvc struct {
	c     *config.Config
	root  *ServiceContext
	repo  *repo.Repo
	seqer *seqer.Client
}

func NewCommentSvc(ctx *ServiceContext, repo *repo.Repo) *CommentSvc {
	s := &CommentSvc{
		c:    ctx.Config,
		repo: repo,
		root: ctx,
	}

	var err error
	s.seqer, err = seqer.NewClient(seqer.WithGrpc(s.c.Seqer.Addr))
	if err != nil {
		panic(err)
	}

	return s
}

func isRootReply(root, parent uint64) bool {
	return root == 0 && parent == 0
}

// 发表评论
func (s *CommentSvc) ReplyAdd(ctx context.Context, req *model.ReplyReq) (*model.ReplyRes, error) {
	var (
		uid      = metadata.Uid(ctx)
		oid      = req.Oid
		rootId   = req.RootId
		parentId = req.ParentId
		ip       = xnet.IpAsInt(metadata.ClientIp(ctx))
	)

	// 取一个号
	replyId, err := s.seqer.GetId(ctx, seqerReplyKey, 10000)
	if err != nil {
		logx.Errorf("reply add get reply id err: %v", err)
		return nil, global.ErrInternal
	}

	now := time.Now().Unix()
	reply := comm.Model{
		Id:       replyId,
		Oid:      oid,
		CType:    int8(req.Type),
		Content:  req.Content,
		Uid:      uid,
		RootId:   rootId,
		ParentId: parentId,
		ReplyUid: req.ReplyUid,
		State:    int8(model.ReplyStateNormal),
		Ip:       ip,
		Ctime:    now,
		Mtime:    now,
	}

	if isRootReply(rootId, parentId) {
		err = s.repo.Bus.AddReply(ctx, &reply)
		if err != nil {
			logx.Errorf("push reply to queue err: %v, replyId: %d", err, replyId)
			return nil, global.ErrInternal
		}
	} else {
		err := s.repo.DB().TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
			// 这里校验非主评论是否能够插入
			if err := s.canAddSubReply(ctx, tx, rootId, parentId); err != nil {
				return err
			}
			// 可以插入
			err = s.repo.Bus.AddReply(ctx, &reply)
			if err != nil {
				logx.Errorf("push subreply to queue err: %v, replyId: %d", err, replyId)
				return global.ErrInternal
			}
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	// TODO notify reply_uid

	return &model.ReplyRes{ReplyId: replyId, Uid: uid}, nil
}

func (s *CommentSvc) canYouDel(ctx context.Context, reply *comm.Model) error {
	var (
		uid   = metadata.Uid(ctx)
		owner = reply.Uid
	)

	if uid == owner {
		// 用户是评论的作者 可以删除
		return nil
	}

	if err := s.userOwnsOid(ctx, uid, reply.Oid, reply.Id); err == nil {
		// 用户是评论对象的作者 可以删除
		return nil
	}

	return global.ErrYouDontOwnThis
}

func (s *CommentSvc) ReplyDel(ctx context.Context, rid uint64) error {
	reply, err := s.repo.CommentRepo.FindById(ctx, rid)
	if err != nil {
		if !xsql.IsNotFound(err) {
			logx.Errorf("reply del find by id err: %v, rid: %d", err, rid)
			return global.ErrInternal
		}
		return global.ErrReplyNotFound
	}

	// 检查用户是否有权限删除评论
	if err := s.canYouDel(ctx, reply); err != nil {
		return err
	}

	err = s.repo.Bus.DelReply(ctx, rid, reply)
	if err != nil {
		logx.Errorf("del reply to queue err: %v, rid: %d", err, rid)
		return global.ErrInternal
	}

	return nil
}

// 点赞
func (s *CommentSvc) ReplyLike(ctx context.Context, rid uint64) error {
	if err := s.repo.Bus.LikeReply(ctx, rid); err != nil {
		logx.Errorf("like reply to queue err: %v, rid: %d", err, rid)
		return global.ErrInternal
	}

	return nil
}

// 取消点赞
func (s *CommentSvc) ReplyUnlike(ctx context.Context, rid uint64) error {
	if err := s.repo.Bus.UnLikeReply(ctx, rid); err != nil {
		logx.Errorf("unlike reply to queue err: %v, rid: %d", err, rid)
		return global.ErrInternal
	}

	return nil
}

// 点踩
func (s *CommentSvc) ReplyDislike(ctx context.Context, rid uint64) error {
	if err := s.repo.Bus.DisLikeReply(ctx, rid); err != nil {
		logx.Errorf("dislike reply to queue err: %v, rid: %d", err, rid)
		return global.ErrInternal
	}

	return nil
}

// 取消点踩
func (s *CommentSvc) ReplyUndislike(ctx context.Context, rid uint64) error {
	if err := s.repo.Bus.UndisLikeReply(ctx, rid); err != nil {
		logx.Errorf("undislike reply to queue err: %v, rid: %d", err, rid)
		return global.ErrInternal
	}

	return nil
}

func (s *CommentSvc) userOwnsOid(ctx context.Context, uid, oid, rid uint64) error {
	resp, err := external.GetNoter().IsUserOwnNote(ctx,
		&notesdk.IsUserOwnNoteReq{
			Uid:    uid,
			NoteId: oid,
		})
	if err != nil {
		logx.Errorf("check IsUserOwnNote err: %v, rid: %d, uid: %d, oid: %d", err, rid, uid, oid)
		return global.ErrInternal
	}

	if !resp.GetResult() {
		return global.ErrYouDontOwnThis
	}

	return nil
}

func (s *CommentSvc) findReplyForUpdate(ctx context.Context, tx sqlx.Session, rid uint64) (*comm.Model, error) {
	m, err := s.repo.CommentRepo.FindByIdForUpdate(ctx, tx, rid)
	if err != nil {
		if !xsql.IsNotFound(err) {
			logx.Errorf("repo find root by id for update err: %v, rid: %d", err, rid)
			return nil, global.ErrInternal
		}
		return nil, global.ErrReplyNotFound
	}

	return m, nil
}

func (s *CommentSvc) canAddSubReply(ctx context.Context, tx sqlx.Session, rootId, parentId uint64) error {
	if rootId != 0 {
		_, err := s.findReplyForUpdate(ctx, tx, rootId)
		if err != nil {
			return err
		}
	}

	if parentId != 0 {
		_, err := s.findReplyForUpdate(ctx, tx, parentId)
		if err != nil {
			return err
		}
	}

	return nil
}

// job related methods
func (s *CommentSvc) ConsumeAddReplyEv(ctx context.Context, data *queue.AddReplyData) error {
	var (
		oid      = data.Oid
		rootId   = data.RootId
		parentId = data.ParentId
	)

	noteExitsRes, err := external.GetNoter().IsNoteExist(ctx,
		&notesdk.IsNoteExistReq{
			NoteId: oid,
		})
	if err != nil {
		logx.Errorf("noter is note exists err: %v, nid: %d", err, oid)
		return err
	}

	// 被评论的对象不存在就直接不操作
	if !noteExitsRes.Exist {
		return nil
	}

	// 新增的是主评论 直接新增
	if isRootReply(rootId, parentId) {
		_, err := s.repo.CommentRepo.Insert(ctx, (*comm.Model)(data))
		if err != nil {
			logx.Errorf("repo insert err: %v, rid: %d", err, data.Id)
			return err
		}
	} else {
		// 新增的是评论的评论 插入前再次校验
		// 检查被评论的平时是否存在
		err := s.repo.DB().TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
			err := s.canAddSubReply(ctx, tx, rootId, parentId)
			if err != nil {
				return err
			}

			// 已经确认被评论的评论与其根评论都存在
			// 可以添加该评论
			_, err = s.repo.CommentRepo.Insert(ctx, (*comm.Model)(data))
			if err != nil {
				logx.Errorf("repo insert err: %v, id: %d", err, data.Id)
				return err
			}

			return nil
		})

		if err != nil {
			logx.Errorf("repo tx all reply err: %v", err)
			return err
		}
	}

	return nil
}

func (s *CommentSvc) ConsumeDelReplyEv(ctx context.Context, data *queue.DelReplyData) error {
	var (
		rid = data.ReplyId
	)

	// 是否是主评论 如果为主评论 需要一并删除所有子评论
	if isRootReply(data.Reply.RootId, data.Reply.ParentId) {
		s.repo.DB().TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
			_, err := s.findReplyForUpdate(ctx, tx, rid)
			if err != nil {
				return err
			}
			// 删除主评论
			err = s.repo.CommentRepo.DeleteByIdTx(ctx, tx, rid)
			if err != nil {
				logx.Errorf("repo delete by id tx err: %v, rid: %d", err, rid)
				return global.ErrInternal
			}
			// 删除旗下子评论
			err = s.repo.CommentRepo.DeleteByRootTx(ctx, tx, rid)
			if err != nil {
				logx.Errorf("repo delete by root tx err: %v, rid: %d", err, rid)
				return global.ErrInternal
			}

			return nil
		})

	} else {
		// 只需要删除评论本身
		err := s.repo.CommentRepo.DeleteById(ctx, rid)
		if err != nil {
			logx.Errorf("repo delete by id err: %v, rid: %d", err, rid)
			return global.ErrInternal
		}
	}
	return nil
}

// 处理点赞或者点踩
// TODO 对接点赞系统
func (s *CommentSvc) ConsumeLikeDislikeEv(ctx context.Context, data *queue.LikeReplyData) error {
	return nil
}

func (s *CommentSvc) PageGetReply(ctx context.Context, in *sdk.PageGetReplyReq) error {
	
	return nil
}

func (s *CommentSvc) PageGetSubReply(ctx context.Context, in *sdk.PageGetSubReplyReq) error {
	
	return nil
}
