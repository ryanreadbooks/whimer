package svc

import (
	"context"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/comment/internal/external"
	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/comment/internal/model"
	"github.com/ryanreadbooks/whimer/comment/internal/repo"
	"github.com/ryanreadbooks/whimer/comment/internal/repo/comm"
	"github.com/ryanreadbooks/whimer/comment/internal/repo/queue"
	commentv1 "github.com/ryanreadbooks/whimer/comment/sdk/v1"
	counterv1 "github.com/ryanreadbooks/whimer/counter/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xnet"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"

	seqer "github.com/ryanreadbooks/folium/sdk"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"golang.org/x/sync/errgroup"
)

const (
	seqerReplyKey = "reply-id-seqer"
)

type CommentSvc struct {
	c     *config.Config
	root  *ServiceContext
	repo  *repo.Repo
	seqer *seqer.Client
	cache *Cache

	// 同步或者异步处理写入数据
	directProxy *commentSvcProxy
	dataProxy   IDataProxy
}

func NewCommentSvc(ctx *ServiceContext, repo *repo.Repo, cache *redis.Redis) *CommentSvc {
	seqer, err := seqer.NewClient(seqer.WithGrpc(ctx.Config.Seqer.Addr))
	if err != nil {
		panic(err)
	}

	s := &CommentSvc{
		c:     ctx.Config,
		root:  ctx,
		repo:  repo,
		seqer: seqer,
		cache: NewCache(cache),
	}

	s.directProxy = &commentSvcProxy{proxy: s}
	s.dataProxy = s.directProxy

	return s
}

func (s *CommentSvc) DataProxy() IDataProxy {
	if s.c.GetDataProxyMode() == global.ProxyModeBus {
		return s.repo.Bus
	}

	return s.directProxy
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

	_, err := external.GetNoter().IsNoteExist(ctx,
		&notev1.IsNoteExistRequest{
			NoteId: oid,
		})
	if err != nil {
		if xerror.ShouldLogError(err) {
			xlog.Msg("noter check note exists err").Err(err).Extra("oid", oid).Errorx(ctx)
		}
		return nil, err
	}

	// 取一个号
	replyId, err := s.seqer.GetId(ctx, seqerReplyKey, 10000)
	if err != nil {
		xlog.Msg("reply add GetId err").Err(err).Errorx(ctx)
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
		err = s.dataProxy.AddReply(ctx, &reply)
		if err != nil {
			xlog.Msg("push reply to queue err").Err(err).Extra("replyId", replyId).Errorx(ctx)
			return nil, err
		}
	} else {
		err := s.repo.DB().TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
			// 这里校验非主评论是否能够插入
			if err := s.canAddSubReply(ctx, tx, rootId, parentId); err != nil {
				return err
			}
			// 可以插入
			err = s.dataProxy.AddReply(ctx, &reply)
			if err != nil {
				xlog.Msg("push subreply to queue err").Err(err).Extra("replyId", replyId).Errorx(ctx)
				return err
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

	if err := s.userOwnsOid(ctx, uid, reply.Oid); err == nil {
		// 用户是评论对象的作者 可以删除
		return nil
	}

	return global.ErrYouDontOwnThis
}

func (s *CommentSvc) ReplyDel(ctx context.Context, rid uint64) error {
	// 检查评论是否存在
	reply, err := s.repo.CommentRepo.FindById(ctx, rid)
	if err != nil {
		if !xsql.IsNotFound(err) {
			xlog.Msg("reply del find by id err").Err(err).Extra("rid", rid).Errorx(ctx)
			return global.ErrInternal
		}
		return global.ErrReplyNotFound
	}

	// 检查用户是否有权限删除评论
	if err := s.canYouDel(ctx, reply); err != nil {
		return err
	}

	err = s.dataProxy.DelReply(ctx, rid, reply)
	if err != nil {
		xlog.Msg("del reply to queue err").Err(err).Extra("rid", rid).Errorx(ctx)
		return err
	}

	return nil
}

// 点赞/取消点赞
func (s *CommentSvc) ReplyLike(ctx context.Context, rid uint64, action int8) error {
	var (
		uid = metadata.Uid(ctx)
	)

	if action == int8(commentv1.ReplyAction_REPLY_ACTION_DO) {
		if err := s.dataProxy.LikeReply(ctx, rid, uid); err != nil {
			xlog.Msg("like reply to queue err").Err(err).Extra("rid", rid).Errorx(ctx)
			return err
		}
		return nil
	}

	if err := s.dataProxy.UnLikeReply(ctx, rid, uid); err != nil {
		xlog.Msg("unlike reply to queue err").Err(err).Extra("rid", rid).Errorx(ctx)
		return err
	}

	return nil
}

// 点踩/取消点踩
func (s *CommentSvc) ReplyDislike(ctx context.Context, rid uint64, action int8) error {
	var (
		uid = metadata.Uid(ctx)
	)

	if action == int8(commentv1.ReplyAction_REPLY_ACTION_DO) {
		if err := s.dataProxy.DisLikeReply(ctx, rid, uid); err != nil {
			xlog.Msg("dislike reply to queue err").Err(err).Extra("rid", rid).Errorx(ctx)
			return err
		}
	}

	if err := s.dataProxy.UnDisLikeReply(ctx, rid, uid); err != nil {
		xlog.Msg("undislike reply to queue err").Err(err).Extra("rid", rid).Errorx(ctx)
		return err
	}

	return nil
}

// 置顶/取消置顶评论
func (s *CommentSvc) ReplyPin(ctx context.Context, oid, rid uint64, action int8) error {
	var (
		uid = metadata.Uid(ctx)
	)

	// 检查rid 不能对非主评论进行操作
	// 找到需要被操作置顶或非置顶的目标评论
	r, err := s.repo.CommentRepo.FindRootParent(ctx, rid)
	if err != nil {
		if xsql.IsNotFound(err) {
			return global.ErrReplyNotFound
		}
		xlog.Msg("repo find uid root parent err").Err(err).
			Extra("rid", rid).Extra("action", action).Errorx(ctx)
		return global.ErrPinFailInternal
	}
	if r.Oid != oid {
		return global.ErrOidNotMatch
	}

	// 检查操作权限 只有oid的作者才能置顶评论
	err = s.userOwnsOid(ctx, uid, r.Oid)
	if err != nil {
		return global.ErrYouCantPinReply
	}

	if !isRootReply(r.RootId, r.ParentId) {
		return global.ErrPinFailNotRoot
	}

	if action == int8(commentv1.ReplyAction_REPLY_ACTION_DO) {
		if r.IsPin == comm.AlreadyPinned {
			return nil
		}
		// 置顶
		err = s.dataProxy.PinReply(ctx, r.Oid, rid)
		if err != nil {
			xlog.Msg("bus put pin reply err").Err(err).
				Extra("rid", rid).Extra("action", action).Errorx(ctx)
			return global.ErrPinFailInternal
		}

	} else {
		// 取消置顶
		if r.IsPin == comm.NotPinned {
			return nil
		}

		err = s.dataProxy.UnPinReply(ctx, r.Oid, rid)
		if err != nil {
			xlog.Msg("bus put unpin reply err").Err(err).
				Extra("rid", rid).Extra("action", action).Errorx(ctx)
			return global.ErrUnPinFailInternal
		}
	}

	return nil
}

func (s *CommentSvc) userOwnsOid(ctx context.Context, uid, oid uint64) error {
	resp, err := external.GetNoter().IsUserOwnNote(ctx,
		&notev1.IsUserOwnNoteRequest{
			Uid:    uid,
			NoteId: oid,
		})
	if err != nil {
		xlog.Msg("check IsUserOwnNote err").Err(err).Extra("oid", oid).Errorx(ctx)
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
			xlog.Msg("repo find root by id for update err").Err(err).Extra("rid", rid).Errorx(ctx)
			return nil, global.ErrInternal
		}
		return nil, global.ErrReplyNotFound
	}

	return m, nil
}

// 检查是否能够发布子评论
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
		&notev1.IsNoteExistRequest{
			NoteId: oid,
		})

	if err != nil {
		if xerror.ShouldLogError(err) {
			xlog.Msg("noter check note exists err").Err(err).Extra("oid", oid).Errorx(ctx)
		}
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
			xlog.Msg("repo insert err").Err(err).Extra("rid", data.Id).Errorx(ctx)
			return err
		}
	} else {
		// 新增的是评论的评论 插入前再次校验
		// 检查被评论的评论是否存在
		err := s.repo.DB().TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
			err := s.canAddSubReply(ctx, tx, rootId, parentId)
			if err != nil {
				return err
			}

			// 已经确认被评论的评论与其根评论都存在
			// 可以添加该评论
			_, err = s.repo.CommentRepo.Insert(ctx, (*comm.Model)(data))
			if err != nil {
				xlog.Msg("repo insert err").Err(err).Extra("rid", data.Id).Errorx(ctx)
				return err
			}

			return nil
		})

		if err != nil {
			xlog.Msg("repo tx all reply err").Err(err).Errorx(ctx)
			return err
		}
	}

	// 更新评论数量
	err = s.cache.IncrReplyCountWhenExist(ctx, oid, 1)
	if err != nil && err != redis.Nil {
		xlog.Msg("cache incr reply count failed").Err(err).Errorx(ctx)
	}

	return nil
}

func (s *CommentSvc) ConsumeDelReplyEv(ctx context.Context, data *queue.DelReplyData) error {
	var (
		rid = data.ReplyId
	)

	// 是否是主评论 如果为主评论 需要一并删除所有子评论
	if isRootReply(data.Reply.RootId, data.Reply.ParentId) {
		err := s.repo.DB().TransactCtx(ctx, func(ctx context.Context, tx sqlx.Session) error {
			_, err := s.findReplyForUpdate(ctx, tx, rid)
			if err != nil {
				return err
			}
			// 删除主评论
			err = s.repo.CommentRepo.DeleteByIdTx(ctx, tx, rid)
			if err != nil {
				xlog.Msg("repo delete by id tx err").Err(err).Extra("rid", rid).Errorx(ctx)
				return global.ErrInternal
			}
			// 删除旗下子评论
			err = s.repo.CommentRepo.DeleteByRootTx(ctx, tx, rid)
			if err != nil {
				xlog.Msg("repo delete by root tx err").Err(err).Extra("rid", rid).Errorx(ctx)
				return global.ErrInternal
			}

			return nil
		})
		if err != nil {
			xlog.Msg("repo transact del root failed").Err(err).Extra("rid", rid).Errorx(ctx)
		}
	} else {
		// 只需要删除评论本身
		err := s.repo.CommentRepo.DeleteById(ctx, rid)
		if err != nil {
			xlog.Msg("repo delete by id err").Err(err).Extra("rid", rid).Errorx(ctx)
			return global.ErrInternal
		}
	}

	// 更新评论数量
	err := s.cache.DecrReplyCountWhenExist(ctx, data.Reply.Oid, 1)
	if err != nil && errors.Is(err, redis.Nil) {
		xlog.Msg("cache incr reply count failed").Err(err).Errorx(ctx)
	}

	return nil
}

// 处理点赞或者点踩
func (s *CommentSvc) ConsumeLikeDislikeEv(ctx context.Context, data *queue.BinaryReplyData) error {
	var (
		rid    = data.ReplyId
		uid    = data.Uid
		action = data.Action
		typ    = data.Type
	)
	ctx = metadata.WithUid(ctx, uid)

	var bizcode int32
	if typ == queue.LikeType {
		bizcode = global.CommentLikeBizcode
	} else {
		bizcode = global.CommentDislikeBizcode
	}

	var (
		err error
	)
	if action == queue.ActionDo {
		// add record
		_, err = external.GetCounter().AddRecord(ctx, &counterv1.AddRecordRequest{
			BizCode: bizcode,
			Uid:     uid,
			Oid:     rid,
		})
	} else {
		// cancel record
		_, err = external.GetCounter().CancelRecord(ctx, &counterv1.CancelRecordRequest{
			BizCode: bizcode,
			Uid:     uid,
			Oid:     rid,
		})
	}

	if err != nil {
		xlog.Msg("counter operates record failed").
			Extra("rid", rid).
			Extra("uid", uid).
			Extra("bizcode", bizcode).
			Extra("action", data.Action).
			Extra("type", data.Type).
			Err(err).
			Errorx(ctx)
	}

	return err
}

// 置顶或者取消置顶
// 每个对象仅支持一条置顶评论，后置顶的评论会替代旧的置顶评论的置顶状态
func (s *CommentSvc) ConsumePinEv(ctx context.Context, data *queue.PinReplyData) error {
	rid := data.ReplyId
	oid := data.Oid

	defer func() {
		// 删除缓存
		if err := s.cache.DelPinned(ctx, oid); err != nil {
			xlog.Msg("del pinned failed").Err(err).Extra("oid", oid).Errorx(ctx)
		}
	}()

	if data.Action == queue.ActionDo {
		err := s.repo.CommentRepo.DoPin(ctx, oid, rid)
		if err != nil {
			xlog.Msg("consume repo do pin err").Err(err).Extra("rid", rid).Extra("oid", oid).Errorx(ctx)
			return global.ErrPinFailInternal
		}
	} else {
		// 取消置顶
		err := s.repo.CommentRepo.SetUnPin(ctx, rid)
		if err != nil {
			xlog.Msg("consume repo set unpin err").Err(err).Extra("rid", rid).Extra("oid", oid).Errorx(ctx)
			return global.ErrUnPinFailInternal
		}
	}

	return nil
}

// 获取根评论
func (s *CommentSvc) PageGetReply(ctx context.Context, in *commentv1.PageGetReplyReq) (*commentv1.PageGetReplyRes, error) {
	const (
		want = 10
	)

	data, err := s.repo.CommentRepo.GetRootReplies(ctx, in.Oid, in.Cursor, want)
	if err != nil {
		xlog.Msg("repo get root reply err").Err(err).
			Extra("cursor", in.Cursor).Extra("oid", in.Oid).Errorx(ctx)
		return nil, global.ErrInternal
	}

	dataLen := len(data)
	var nextCursor uint64 = 0
	hasNext := dataLen == want
	if dataLen > 0 {
		nextCursor = data[dataLen-1].Id
	}

	replies := make([]*commentv1.ReplyItem, 0, dataLen)
	for _, item := range data {
		replies = append(replies, modelToReplyItem(item))
	}

	if err := s.fillReplyLikes(ctx, replies); err != nil {
		xlog.Msg("page get reply fill reply likes failed").
			Err(err).Errorx(ctx)
	}

	return &commentv1.PageGetReplyRes{
		Replies:    replies,
		NextCursor: nextCursor,
		HasNext:    hasNext,
	}, nil
}

// 获取子评论
func (s *CommentSvc) PageGetSubReply(ctx context.Context, in *commentv1.PageGetSubReplyReq) (*commentv1.PageGetSubReplyRes, error) {
	const (
		want = 5
	)

	data, err := s.repo.CommentRepo.GetSubReply(ctx, in.Oid, in.RootId, in.Cursor, want)
	if err != nil {
		xlog.Msg("repo get sub reply err").Err(err).
			Extra("cursor", in.Cursor).Extra("oid", in.Oid).Errorx(ctx)
		return nil, global.ErrInternal
	}

	dataLen := len(data)
	var nextCursor uint64 = 0
	hasNext := dataLen == want
	if dataLen > 0 {
		nextCursor = data[dataLen-1].Id
	}
	replies := make([]*commentv1.ReplyItem, 0, dataLen)
	for _, item := range data {
		replies = append(replies, modelToReplyItem(item))
	}

	if err := s.fillReplyLikes(ctx, replies); err != nil {
		xlog.Msg("page get sub reply fill reply likes failed").
			Err(err).Errorx(ctx)
	}

	return &commentv1.PageGetSubReplyRes{
		Replies:    replies,
		NextCursor: nextCursor,
		HasNext:    hasNext,
	}, nil
}

func modelToReplyItem(model *comm.Model) *commentv1.ReplyItem {
	out := &commentv1.ReplyItem{}
	if model == nil {
		return out
	}

	out.Id = model.Id
	out.Oid = model.Oid
	out.ReplyType = uint32(model.CType)
	out.Content = model.Content
	out.Uid = model.Uid
	out.RootId = model.RootId
	out.ParentId = model.ParentId
	out.Ruid = model.ReplyUid
	out.LikeCount = uint64(model.Like)
	out.HateCount = uint64(model.Dislike)
	out.Ctime = model.Ctime
	out.Mtime = model.Mtime
	out.Ip = xnet.IntAsIp(uint32(model.Ip))
	if model.IsPin == 1 {
		out.IsPin = true
	}

	return out
}

func isRootReply(root, parent uint64) bool {
	return root == 0 && parent == 0
}

// 获取评论信息，包含主评论和子评论
func (s *CommentSvc) PageGetObjectReplies(ctx context.Context, in *commentv1.PageGetReplyReq) (*commentv1.PageGetDetailedReplyRes, error) {
	const (
		// 默认拿10条主评论 每条主评论又取其5条子评论
		wantRoot = 10
		wantSub  = 5
	)

	// 先获取主评论
	roots, err := s.PageGetReply(ctx, in)
	if err != nil {
		xlog.Msg("repo get object replies err").Err(err).
			Extra("cursor", in.Cursor).Extra("oid", in.Oid).Errorx(ctx)
		return nil, global.ErrInternal
	}

	replies, err := s.getSubrepliesForRoots(ctx, in.Oid, roots.Replies)
	if err != nil {
		return nil, err
	}

	repliesItems := make([]*commentv1.ReplyItem, 0, len(replies))
	for _, reply := range replies {
		for _, sub := range reply.Subreplies.Items {
			repliesItems = append(repliesItems, sub)
		}
	}

	if err := s.fillReplyLikes(ctx, repliesItems); err != nil {
		xlog.Msg("page get object replies fill reply likes failed").
			Err(err).Errorx(ctx)
	}

	return &commentv1.PageGetDetailedReplyRes{
		Replies:    replies,
		NextCursor: roots.NextCursor,
		HasNext:    roots.HasNext,
	}, nil
}

// 获取roots主评论的子评论
// 并且将子评论和主评论拼接后返回
func (s *CommentSvc) getSubrepliesForRoots(ctx context.Context,
	oid uint64,
	roots []*commentv1.ReplyItem) ([]*commentv1.DetailedReplyItem, error) {

	// 起协程去拿每个主评论的子评论
	wg, ctx := errgroup.WithContext(ctx)
	var subs = make([]*commentv1.PageGetSubReplyRes, len(roots))
	for i, root := range roots {
		idx, rootTmp := i, root
		wg.Go(func() error {
			// 按照oid和root获取子评论
			subItem, err := s.PageGetSubReply(ctx, &commentv1.PageGetSubReplyReq{
				Oid:    oid,
				RootId: rootTmp.Id,
				Cursor: 0,
			})
			if err != nil {
				return err
			}
			subs[idx] = subItem
			return nil
		})
	}

	err := wg.Wait()
	if err != nil {
		xlog.Msg("parallel repo get sub reply err").Err(err).Extra("oid", oid).Errorx(ctx)
		return nil, global.ErrInternal
	}

	// 拼装结果
	replies := make([]*commentv1.DetailedReplyItem, 0, len(roots))
	for i, root := range roots {
		sub := subs[i]
		replies = append(replies, &commentv1.DetailedReplyItem{
			Root: root,
			Subreplies: &commentv1.DetailedSubReply{
				Items:      sub.Replies,
				HasNext:    sub.HasNext,
				NextCursor: sub.NextCursor,
			},
		})
	}

	return replies, nil
}

func (s *CommentSvc) GetPinnedReply(ctx context.Context, oid uint64) (*commentv1.GetPinnedReplyRes, error) {
	// 先找出置顶评论
	root, err := s.cache.GetPinned(ctx, oid)
	if err != nil {
		xlog.Msg("cache get pinned failed").Err(err).Extra("oid", oid).Errorx(ctx)
		root, err = s.repo.CommentRepo.GetPinned(ctx, oid)
		if err != nil {
			if xsql.IsNotFound(err) {
				return &commentv1.GetPinnedReplyRes{}, nil
			}
			xlog.Msg("repo get pinned err").Err(err).Extra("oid", oid).Errorx(ctx)
			return nil, global.ErrGetPinnedInternal
		}

		// set cache
		concurrent.SafeGo(func() {
			ctxc := context.WithoutCancel(ctx)
			err = s.cache.SetPinned(ctxc, root)
			if err != nil {
				xlog.Msg("cache set pinned failed").Err(err).Extra("oid", oid).Errorx(ctxc)
			}
		})
	}

	// 随后找出置顶评论的子评论
	rootWithSubs, err := s.getSubrepliesForRoots(ctx, oid, []*commentv1.ReplyItem{modelToReplyItem(root)})
	if err != nil {
		return nil, err
	}

	replies := make([]*commentv1.ReplyItem, 0)
	replies = append(replies, rootWithSubs[0].Root)
	replies = append(replies, rootWithSubs[0].Subreplies.Items...)
	if err := s.fillReplyLikes(ctx, replies); err != nil {
		xlog.Msg("get pinned reply fill reply likes failed").
			Err(err).
			Errorx(ctx)
	}

	return &commentv1.GetPinnedReplyRes{
		Reply: rootWithSubs[0],
	}, nil
}

// 获取被评论对象oid的评论数量
func (s *CommentSvc) CountReply(ctx context.Context, oid uint64) (uint64, error) {
	// fetch from cache
	count, err := s.cache.GetReplyCount(ctx, oid)
	if err != nil {
		xlog.Msg("cache get count failed").Err(err).Extra("oid", oid).Errorx(ctx)
		// fetch from db instead
		count, err = s.repo.CommentRepo.CountByOid(ctx, oid)
		if err != nil {
			xlog.Msg("repo get count failed").Err(err).Extra("oid", oid).Errorx(ctx)
			return 0, global.ErrCountReplyInternal
		}
		err = s.cache.SetReplyCount(ctx, oid, count)
		if err != nil {
			xlog.Msg("cache set reply count failed").Err(err).
				Extra("oid", oid).
				Extra("count", count).
				Errorx(ctx)
		}
	}

	return count, nil
}

// 获取评论点赞数量
func (s *CommentSvc) GetReplyLikesCount(ctx context.Context, rid uint64) (uint64, error) {
	return s.counterGetCount(ctx, rid, global.CommentLikeBizcode)
}

// 获取评论点踩数
func (s *CommentSvc) GetReplyDislikesCount(ctx context.Context, rid uint64) (uint64, error) {
	return s.counterGetCount(ctx, rid, global.CommentDislikeBizcode)
}

// 从counter获取评论点赞/点踩计数
func (s *CommentSvc) counterGetCount(ctx context.Context, rid uint64, biz int32) (uint64, error) {
	summary, err := external.GetCounter().
		GetSummary(ctx, &counterv1.GetSummaryRequest{
			BizCode: biz,
			Oid:     rid,
		})
	if err != nil {
		xlog.Msg("counter get count failed").Err(err).Errorx(ctx)
		if biz == global.CommentLikeBizcode {
			return 0, global.ErrGetReplyLikeCount
		}
		return 0, global.ErrGetReplyDislikeCount
	}

	return summary.Count, nil
}

// 获取用户是否评论过某个对象
func (s *CommentSvc) CheckUserCommentOnObject(ctx context.Context, uid, oid uint64) (bool, error) {
	cnt, err := s.repo.CommentRepo.CountByOidUid(ctx, oid, uid)
	if err != nil {
		return false, xerror.Wrapf(err, "comment repo count by oid and uid failed").
			WithExtras("uid", uid, "oid", oid).
			WithCtx(ctx)
	}

	return cnt != 0, nil
}

// 全量同步评论数量
func (s *CommentSvc) FullSyncReplyCountCache(ctx context.Context) error {
	res, err := s.repo.CommentRepo.CountGroupByOid(ctx)
	if err != nil {
		xlog.Msg("full sync count repo failed").Err(err).Errorx(ctx)
		return err
	}

	err = s.cache.BatchSetReplyCount(ctx, res)
	if err != nil {
		xlog.Msg("full sync set cache failed").Err(err).Errorx(ctx)
		return err
	}

	return nil
}

func (s *CommentSvc) PartialSyncReplyCountCache(ctx context.Context, offset, limit int64) error {
	res, err := s.repo.CommentRepo.CountGroupByOidLimit(ctx, offset, limit)
	if err != nil {
		xlog.Msg("full sync count repo limit failed").Err(err).Errorx(ctx)
		return err
	}

	err = s.cache.BatchSetReplyCount(ctx, res)
	if err != nil {
		xlog.Msg("full sync set cache failed").Err(err).Errorx(ctx)
		return err
	}

	return nil
}
