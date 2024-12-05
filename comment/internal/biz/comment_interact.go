package biz

import (
	"context"

	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/comment/internal/infra"
	"github.com/ryanreadbooks/whimer/comment/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/comment/internal/model"
	counterv1 "github.com/ryanreadbooks/whimer/counter/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

const (
	ActionUndo = 0 // 取消执行操作，比如取消点赞，取消置顶等
	ActionDo   = 1 // 执行操作，比如点赞，置顶等
)

// 评论交互领域 点赞 点踩 置顶 举报等
type CommentInteractBiz interface {
	// 用户点赞/取消点赞评论
	LikeReply(ctx context.Context, rid uint64, action int8) error
	// 用户点踩/取消点踩评论
	DislikeReply(ctx context.Context, rid uint64, action int8) error
	// 用户置顶/取消置顶评论
	PinReply(ctx context.Context, oid, rid uint64, action int8) error
	// 获取评论点赞数量
	CountReplyLikes(ctx context.Context, rid uint64) (uint64, error)
	// 获取评论点踩数量
	// 批量获取评论点赞数量
	BatchCountReplyLikes(ctx context.Context, rids []uint64) (map[uint64]uint64, error)
	// 获取评论点踩数量
	CountReplyDislikes(ctx context.Context, rid uint64) (uint64, error)
	// 批量获取评论点踩数量
	BatchCountReplyDislikes(ctx context.Context, rids []uint64) (map[uint64]uint64, error)
	// 填充评论的点赞数量
	PopulateLike(ctx context.Context, reply *model.ReplyItem) error
	// 填充评论的点赞数量
	PopulateLikes(ctx context.Context, replies []*model.ReplyItem) error
}

type commentInteractBiz struct{}

func NewCommentInteractBiz() CommentInteractBiz {
	return &commentInteractBiz{}
}

// 用户点赞/取消点赞评论
func (b *commentInteractBiz) LikeReply(ctx context.Context, rid uint64, action int8) error {
	err := b.likeOrDislike(ctx, rid, action, global.CommentLikeBizcode)
	if err != nil {
		return xerror.Wrapf(err, "comment interact biz failed to like reply using counter").
			WithExtras("rid", rid, "action", action).WithCtx(ctx)
	}

	return nil
}

// 用户点踩/取消点踩评论
func (b *commentInteractBiz) DislikeReply(ctx context.Context, rid uint64, action int8) error {
	err := b.likeOrDislike(ctx, rid, action, global.CommentDislikeBizcode)
	if err != nil {
		return xerror.Wrapf(err, "comment interact biz failed to dislike reply using counter").
			WithExtras("rid", rid, "action", action).WithCtx(ctx)
	}

	return nil
}

func (b *commentInteractBiz) likeOrDislike(ctx context.Context, rid uint64, action int8, bizcode int32) error {
	var (
		uid = metadata.Uid(ctx)
		err error
	)

	if action == ActionDo {
		// add record
		_, err = dep.GetCounter().AddRecord(ctx, &counterv1.AddRecordRequest{
			BizCode: bizcode,
			Uid:     uid,
			Oid:     rid,
		})
	} else {
		// 取消点赞
		_, err = dep.GetCounter().CancelRecord(ctx, &counterv1.CancelRecordRequest{
			BizCode: bizcode,
			Uid:     uid,
			Oid:     rid,
		})
	}
	if err != nil {
		return xerror.Wrapf(err, "comment interact biz likeOrDislike failed")
	}

	return nil
}

// 用户置顶/取消置顶评论
// 每个对象仅支持一条置顶评论，后置顶的评论会替代旧的置顶评论的置顶状态
func (b *commentInteractBiz) PinReply(ctx context.Context, oid, rid uint64, action int8) error {
	var (
		err error
	)

	if action == ActionDo {
		// 置顶
		err = infra.Dao().CommentDao.DoPin(ctx, oid, rid)
	} else {
		// 取消置顶
		err = infra.Dao().CommentDao.SetUnPin(ctx, rid, oid)
	}

	if err != nil {
		return xerror.Wrapf(err, "comment interact biz pin reply failed").
			WithExtras("oid", oid, "rid", rid, "action", action).
			WithCtx(ctx)
	}

	return nil
}

func (b *commentInteractBiz) getLikeOrDislikeCount(ctx context.Context, rid uint64, bizcode int32) (uint64, error) {
	summary, err := dep.GetCounter().GetSummary(ctx, &counterv1.GetSummaryRequest{
		BizCode: bizcode,
		Oid:     rid,
	})
	if err != nil {
		return 0, xerror.Wrapf(err, "comment interact biz counter get summary failed").WithCtx(ctx)
	}

	return summary.Count, nil
}

func (b *commentInteractBiz) batchGetLikeOrDislikeCount(ctx context.Context, rids []uint64, bizcode int32) (map[uint64]uint64, error) {
	requests := make([]*counterv1.GetSummaryRequest, 0, len(rids))
	for _, r := range rids {
		requests = append(requests, &counterv1.GetSummaryRequest{
			BizCode: bizcode,
			Oid:     r,
		})
	}

	summaries, err := dep.GetCounter().BatchGetSummary(ctx,
		&counterv1.BatchGetSummaryRequest{
			Requests: requests,
		})

	if err != nil {
		return nil, xerror.Wrapf(err, "comment interact biz counter batch get summaries failed").WithCtx(ctx)
	}

	var resp = make(map[uint64]uint64, len(summaries.Responses))
	for _, v := range summaries.Responses {
		resp[v.Oid] = v.Count
	}

	return resp, nil
}

// 获取评论点赞数量
func (b *commentInteractBiz) CountReplyLikes(ctx context.Context, rid uint64) (uint64, error) {
	return b.getLikeOrDislikeCount(ctx, rid, global.CommentLikeBizcode)
}

// 批量获取评论点赞数量
func (b *commentInteractBiz) BatchCountReplyLikes(ctx context.Context, rids []uint64) (map[uint64]uint64, error) {
	return b.batchGetLikeOrDislikeCount(ctx, rids, global.CommentLikeBizcode)
}

// 获取评论点踩数量
func (b *commentInteractBiz) CountReplyDislikes(ctx context.Context, rid uint64) (uint64, error) {
	return b.getLikeOrDislikeCount(ctx, rid, global.CommentDislikeBizcode)
}

// 批量获取评论点踩数量
func (b *commentInteractBiz) BatchCountReplyDislikes(ctx context.Context, rids []uint64) (map[uint64]uint64, error) {
	return b.batchGetLikeOrDislikeCount(ctx, rids, global.CommentDislikeBizcode)
}

func (b *commentInteractBiz) PopulateLike(ctx context.Context, reply *model.ReplyItem) error {
	replies := []*model.ReplyItem{reply}
	return b.PopulateLikes(ctx, replies)
}

// 填充评论的点赞数量
func (b *commentInteractBiz) PopulateLikes(ctx context.Context, replies []*model.ReplyItem) error {
	if len(replies) == 0 {
		return nil
	}

	requests := make([]*counterv1.GetSummaryRequest, 0, 16)
	for _, reply := range replies {
		requests = append(requests, &counterv1.GetSummaryRequest{
			BizCode: global.CommentLikeBizcode,
			Oid:     reply.Id,
		})
	}
	resp, err := dep.GetCounter().BatchGetSummary(ctx,
		&counterv1.BatchGetSummaryRequest{
			Requests: requests,
		})
	if err != nil {
		return xerror.Wrapf(err, "comment interact biz counter batch get summary failed").
			WithExtra("len", len(replies)).WithCtx(ctx)
	}

	type key struct {
		BizCode int32
		Oid     uint64
	}

	mapping := make(map[key]uint64, len(resp.Responses))
	for _, item := range resp.Responses {
		mapping[key{item.BizCode, item.Oid}] = item.Count
	}

	for _, reply := range replies {
		k := key{global.CommentLikeBizcode, reply.Id}
		if cnt, ok := mapping[k]; ok {
			reply.LikeCount = cnt
		}
	}

	return nil
}
