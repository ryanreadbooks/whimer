package biz

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/comment/internal/infra"
	"github.com/ryanreadbooks/whimer/comment/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/comment/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/comment/internal/model"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xnet"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
)

// 评论基础功能领域
type CommentBiz interface {
	// 用户发表评论
	AddReply(ctx context.Context, req *model.AddReplyReq) (*model.AddReplyRes, error)
	// 用户删除评论
	DelReply(ctx context.Context, rid uint64) error
	// 检查评论是否存在
	GetReply(ctx context.Context, rid uint64) (*model.ReplyItem, error)
	// 仅获取根评论
	GetRootReplies(ctx context.Context, oid uint64, cursor uint64, want int, sortBy int8) (*model.PageReplies, error)
	// 仅获取子评论
	GetSubReplies(ctx context.Context, oid, rootId uint64, want int, cursor uint64) (*model.PageReplies, error)
	// 获取评论数量
	CountReply(ctx context.Context, oid uint64) (uint64, error)
	// 批量获取评论数量
	BatchCountReply(ctx context.Context, oids []uint64) (map[uint64]uint64, error)
	// 检查用户是否发起了评论
	CheckUserIsReplied(ctx context.Context, uid, oid uint64) (bool, error)
	// 批量检查用户是否发起了评论
	BatchCheckUserIsReplied(ctx context.Context, uidOids map[uint64][]uint64) ([]model.UidCommentOnOid, error)
	// 获取置顶评论
	GetPinnedReply(ctx context.Context, oid uint64) (*model.ReplyItem, error)
}

const (
	seqerReplyKey = "reply-id-seqer"
)

type commentBiz struct{}

func NewCommentBiz() CommentBiz {
	return &commentBiz{}
}

func (b *commentBiz) AddReply(ctx context.Context, req *model.AddReplyReq) (*model.AddReplyRes, error) {
	var (
		uid      = metadata.Uid(ctx)
		oid      = req.Oid
		rootId   = req.RootId
		parentId = req.ParentId
		ip       = xnet.IpAsInt(metadata.ClientIp(ctx))
	)

	// 必须笔记存在才可以添加评论
	noteExitsRes, err := dep.GetNoter().IsNoteExist(ctx,
		&notev1.IsNoteExistRequest{
			NoteId: oid,
		})

	if err != nil {
		return nil, xerror.Wrapf(err, "comment biz check note exists failed").WithCtx(ctx)
	}

	// 被评论的对象不存在就直接不操作
	if !noteExitsRes.Exist {
		return nil, global.ErrNoNote
	}

	// 取一个号
	replyId, err := dep.ReplyIdgen().GetId(ctx, seqerReplyKey, 10000)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment biz failed to gen replyId").WithCtx(ctx)
	}

	now := time.Now().Unix()
	reply := dao.Comment{
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

	// 新增的是主评论 直接新增
	if model.IsRoot(rootId, parentId) {
		_, err := infra.Dao().CommentDao.Insert(ctx, &reply)
		if err != nil {
			return nil, xerror.Wrapf(err, "comment biz insert root reply failed")
		}
	} else {
		// 新增的是评论的评论 插入前校验能否插入
		// 检查被评论的评论是否存在
		err = infra.Dao().DB().Transact(ctx, func(ctx context.Context) error {
			err := b.isReplyAddable(ctx, rootId, parentId)
			if err != nil {
				return xerror.Wrapf(err, "isReplyAddable check failed")
			}

			// 可以插入
			_, err = infra.Dao().CommentDao.Insert(ctx, &reply)
			if err != nil {
				return xerror.Wrapf(err, "comment biz dao insert reply failed")
			}

			return nil
		})
		if err != nil {
			return nil, xerror.Wrapf(err, "comment biz failed to insert reply")
		}
	}

	concurrent.DoneIn(10*time.Second, func(ctx context.Context) {
		infra.Dao().CommentDao.IncrReplyCount(ctx, oid)
	})

	return &model.AddReplyRes{Uid: uid, ReplyId: replyId}, nil
}

func (b *commentBiz) findByIdForUpdate(ctx context.Context, rid uint64) (*model.ReplyItem, error) {
	c, err := infra.Dao().CommentDao.FindByIdForUpdate(ctx, rid)
	if err != nil {
		if !xsql.IsNotFound(err) {
			return nil, xerror.Wrapf(err, "comment biz find by for update failed")
		}
		return nil, xerror.Wrap(global.ErrReplyNotFound)
	}

	return model.NewReplyItem(c), nil
}

// 检查是否能够发布子评论
func (b *commentBiz) isReplyAddable(ctx context.Context, rootId, parentId uint64) error {
	// 两个都需要检查
	if rootId != 0 {
		_, err := b.findByIdForUpdate(ctx, rootId)
		if err != nil {
			return xerror.Wrap(err)
		}
		return nil
	}

	if parentId != 0 && rootId != parentId {
		_, err := b.findByIdForUpdate(ctx, parentId)
		if err != nil {
			return xerror.Wrap(err)
		}
		return nil
	}

	return nil
}

// 用户删除评论
func (b *commentBiz) DelReply(ctx context.Context, rid uint64) error {
	var (
		uid = metadata.Uid(ctx)
	)

	// 检查评论是否存在
	reply, err := b.GetReply(ctx, rid)
	if err != nil {
		return xerror.Wrapf(err, "comment biz failed to get reply")
	}

	if err := b.isReplyDeletable(ctx, uid, reply); err != nil {
		return xerror.Wrapf(err, "comment biz check replys is not deletable")
	}

	// 开始删除
	// 删除的如果是主评论 需要一并删除所有子评论
	// 否则就只删除评论本身
	if model.IsRoot(reply.RootId, reply.ParentId) {
		err = infra.Dao().DB().Transact(ctx, func(ctx context.Context) error {
			// 上锁
			_, err := b.findByIdForUpdate(ctx, rid)
			if err != nil {
				return xerror.Wrapf(err, "comment biz dao find reply for update failed")
			}
			// 删除主评论
			err = infra.Dao().CommentDao.DeleteById(ctx, rid)
			if err != nil {
				return xerror.Wrapf(err, "comment biz dao delete root by id failed")
			}

			// 删除其下子评论
			err = infra.Dao().CommentDao.DeleteByRoot(ctx, rid)
			if err != nil {
				return xerror.Wrapf(err, "comment biz dao delete replies by rootid failed")
			}

			return nil
		})
		if err != nil {
			return xerror.Wrapf(err, "comment biz failed to delete reply")
		}
	} else {
		if err = infra.Dao().CommentDao.DeleteById(ctx, rid); err != nil {
			return xerror.Wrapf(err, "comment biz failed to delete reply")
		}
	}

	concurrent.DoneIn(10*time.Second, func(ctx context.Context) {
		infra.Dao().CommentDao.DecrReplyCount(ctx, reply.Oid)
	})
	return nil
}

func (b *commentBiz) GetReply(ctx context.Context, rid uint64) (*model.ReplyItem, error) {
	reply, err := infra.Dao().CommentDao.FindById(ctx, rid)
	if err != nil {
		if !xsql.IsNotFound(err) {
			return nil, xerror.Wrapf(err, "comment biz failed to get reply").WithExtra("rid", rid).WithCtx(ctx)
		}

		return nil, xerror.Wrap(global.ErrReplyNotFound)
	}

	return model.NewReplyItem(reply), nil
}

// 检查是否可以删除评论, 比如用户权限判断
func (b *commentBiz) isReplyDeletable(ctx context.Context, uid uint64, reply *model.ReplyItem) error {
	var (
		owner = reply.Uid
	)

	// 用户是评论的作者可以删除
	if uid == owner {
		return nil
	}

	// 用户是评论对象的作者可以删除
	resp, err := dep.GetNoter().IsUserOwnNote(ctx, &notev1.IsUserOwnNoteRequest{
		Uid:    uid,
		NoteId: reply.Oid,
	})
	if err != nil {
		return xerror.Wrapf(err, "comment biz failed to check owner").WithExtra("replyId", reply.Oid).WithCtx(ctx)
	}

	if !resp.GetResult() {
		return xerror.Wrap(global.ErrYouDontOwnThis)
	}

	return nil
}

// 仅获取根评论, 不获取根评论下的子评论
// 每次返回10条
func (b *commentBiz) GetRootReplies(ctx context.Context, oid uint64, cursor uint64, want int, sortBy int8) (*model.PageReplies, error) {
	if want <= 0 {
		want = 10
	}

	data, err := infra.Dao().CommentDao.GetRootReplies(ctx, oid, cursor, want)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment biz failed to get root replies").WithCtx(ctx).
			WithExtras("oid", oid, "cursor", cursor)
	}

	// 计算下一个cursor
	dataLen := len(data)
	var nextCursor uint64 = 0
	hasNext := dataLen == want
	if dataLen > 0 {
		nextCursor = data[dataLen-1].Id
	}

	items := make([]*model.ReplyItem, 0, dataLen)
	for _, item := range data {
		items = append(items, model.NewReplyItem(item))
	}

	return &model.PageReplies{
		Items:      items,
		NextCursor: nextCursor,
		HasNext:    hasNext,
	}, nil
}

// 仅获取子评论
// 获取对象oid中rootId评论下的子评论
// 每次返回5条
func (b *commentBiz) GetSubReplies(ctx context.Context, oid, rootId uint64, want int, cursor uint64) (*model.PageReplies, error) {
	if want <= 0 {
		want = 5
	}

	data, err := infra.Dao().CommentDao.GetSubReplies(ctx, oid, rootId, cursor, want)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment biz failed to get sub replies").
			WithExtras("oid", oid, "rootId", rootId, "cursor", cursor).WithCtx(ctx)
	}
	dataLen := len(data)
	var nextCursor uint64 = 0
	hasNext := dataLen == want
	if dataLen > 0 {
		nextCursor = data[dataLen-1].Id
	}

	items := make([]*model.ReplyItem, 0, dataLen)
	for _, item := range data {
		items = append(items, model.NewReplyItem(item))
	}

	return &model.PageReplies{
		Items:      items,
		NextCursor: nextCursor,
		HasNext:    hasNext,
	}, nil
}

func (b *commentBiz) CountReply(ctx context.Context, oid uint64) (uint64, error) {
	cnt, err := infra.Dao().CommentDao.CountByOid(ctx, oid)
	if err != nil {
		return 0, xerror.Wrapf(err, "comment biz failed to get replies count").WithCtx(ctx).WithExtra("oid", oid)
	}

	return cnt, nil
}

// 批量获取评论数量
func (b *commentBiz) BatchCountReply(ctx context.Context, oids []uint64) (map[uint64]uint64, error) {
	resp, err := infra.Dao().CommentDao.BatchCountByOid(ctx, oids)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment biz failed to batch get replies count").WithCtx(ctx)
	}

	return resp, nil
}

// 检查用户是否发起了评论
func (b *commentBiz) CheckUserIsReplied(ctx context.Context, uid, oid uint64) (bool, error) {
	cnt, err := infra.Dao().CommentDao.CountByOidUid(ctx, oid, uid)
	if err != nil {
		return false, xerror.Wrapf(err, "comment biz failed to check user is replied object").
			WithExtras("uid", uid, "oid", oid).
			WithCtx(ctx)
	}

	return cnt != 0, nil
}

// 批量检查用户是否发起了评论
// uid => [oid1, oid2, ..., oidN]
func (b *commentBiz) BatchCheckUserIsReplied(ctx context.Context, uidOids map[uint64][]uint64) ([]model.UidCommentOnOid, error) {
	resp, err := infra.Dao().CommentDao.FindByUidsOids(ctx, uidOids)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment biz failed to batch check user is replied").WithCtx(ctx)
	}

	// 记录uid评论过哪些oids
	commented := make(map[uint64][]uint64, len(resp))
	for _, r := range resp {
		commented[r.Uid] = append(commented[r.Uid], r.Oid)
	}

	var result = make([]model.UidCommentOnOid, 0, len(uidOids))
	// commented和req的进行对比 得到req中uid是否评论某oid
	for uid, targets := range uidOids {
		oidCommenteds := commented[uid]
		for _, oidChecked := range targets {
			cmted := false
			for _, oidCmted := range oidCommenteds {
				if oidChecked == oidCmted {
					cmted = true
				}
			}
			result = append(result, model.UidCommentOnOid{
				Uid:       uid,
				Oid:       oidChecked,
				Commented: cmted,
			})
		}
	}

	return result, nil
}

func (b *commentBiz) GetPinnedReply(ctx context.Context, oid uint64) (*model.ReplyItem, error) {
	pinned, err := infra.Dao().CommentDao.GetPinned(ctx, oid)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment biz get pinned reply failed").WithExtra("oid", oid).WithCtx(ctx)
	}

	return model.NewReplyItem(pinned), nil
}
