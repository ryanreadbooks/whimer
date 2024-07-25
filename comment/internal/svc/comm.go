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
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xnet"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	notesdk "github.com/ryanreadbooks/whimer/note/sdk"

	seqer "github.com/ryanreadbooks/folium/sdk"
	"github.com/zeromicro/go-zero/core/logx"
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
	s.seqer, err = seqer.NewClient(seqer.WithGrpc(s.c.External.Grpc.Seqer))
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

	err = s.repo.Bus.AddReply(ctx, &reply)
	if err != nil {
		logx.Errorf("push reply to queue err: %v, replyId: %d", err, replyId)
		return nil, global.ErrInternal
	}

	// TODO notify reply_uid

	return &model.ReplyRes{ReplyId: replyId, Uid: uid}, nil
}

func (s *CommentSvc) ReplyDel(ctx context.Context, rid uint64) error {
	var (
		uid = metadata.Uid(ctx)
	)

	reply, err := s.repo.CommentRepo.FindById(ctx, rid)
	if err != nil {
		if !xsql.IsNotFound(err) {
			logx.Errorf("reply del find by id err: %v, rid: %d", err, rid)
			return global.ErrInternal
		}
		return global.ErrReplyNotFound
	}

	// 检查用户是否有权删除该评论，如下情况之一可以删
	// 1. 用户是该评论的作者
	// 2. 用户是该评论对象的作者

	if uid != reply.Uid {
		resp, err := external.GetNoter().IsUserOwnNote(ctx, &notesdk.IsUserOwnNoteReq{
			Uid:     uid,
			NoteIds: []uint64{reply.Oid},
		})
		if err != nil {
			logx.Errorf("check IsUserOwnNote err: %v, rid: %d, uid: %d", err, rid, uid)
			return global.ErrInternal
		}

		if len(resp.GetResult()) < 1 {
			logx.Errorf("check IsUserOwnNote result len is 0: rid: %d, uid: %d", rid, uid)
			return global.ErrInternal
		}

		if !resp.GetResult()[0] {
			return global.ErrPermDenied
		}
	}

	// 是否是主评论 如果为主评论 需要一并删除所有子评论
	if isRootReply(reply.RootId, reply.ParentId) {

	} else {
		// 只需要删除评论本身
		
	}

	return nil
}
