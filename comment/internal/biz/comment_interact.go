package biz

import (
	"context"

	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/comment/internal/infra"
	"github.com/ryanreadbooks/whimer/comment/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/comment/internal/model"
	counterv1 "github.com/ryanreadbooks/whimer/counter/api/v1"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

const (
	ActionUndo = 0 // 取消执行操作，比如取消点赞，取消置顶等
	ActionDo   = 1 // 执行操作，比如点赞，置顶等
)

type CommentInteractBiz struct{}

// 评论交互领域 点赞 点踩 置顶 举报等
func NewCommentInteractBiz() CommentInteractBiz {
	return CommentInteractBiz{}
}

// 用户点赞/取消点赞评论
func (b *CommentInteractBiz) LikeComment(ctx context.Context, commentId int64, action int8) error {
	err := b.likeOrDislike(ctx, commentId, action, global.CommentLikeBizcode)
	if err != nil {
		return xerror.Wrapf(err, "comment interact biz failed to like comment using counter").
			WithExtras("cid", commentId, "action", action).WithCtx(ctx)
	}

	return nil
}

// 用户点踩/取消点踩评论
func (b *CommentInteractBiz) DislikeComment(ctx context.Context, commentId int64, action int8) error {
	err := b.likeOrDislike(ctx, commentId, action, global.CommentDislikeBizcode)
	if err != nil {
		return xerror.Wrapf(err, "comment interact biz failed to dislike comment using counter").
			WithExtras("cid", commentId, "action", action).WithCtx(ctx)
	}

	return nil
}

func (b *CommentInteractBiz) likeOrDislike(ctx context.Context, commentId int64, action int8, bizcode int32) error {
	var (
		uid = metadata.Uid(ctx)
		err error
	)

	if action == ActionDo {
		// add record
		_, err = dep.GetCounter().AddRecord(ctx, &counterv1.AddRecordRequest{
			BizCode: bizcode,
			Uid:     uid,
			Oid:     commentId,
		})
	} else {
		// 取消点赞
		_, err = dep.GetCounter().CancelRecord(ctx, &counterv1.CancelRecordRequest{
			BizCode: bizcode,
			Uid:     uid,
			Oid:     commentId,
		})
	}
	if err != nil {
		return xerror.Wrapf(err, "comment interact biz likeOrDislike failed")
	}

	return nil
}

// 用户置顶/取消置顶评论
// 每个对象仅支持一条置顶评论，后置顶的评论会替代旧的置顶评论的置顶状态
func (b *CommentInteractBiz) PinComment(ctx context.Context, oid, commentId int64, action int8) error {
	var (
		err error
	)

	if action == ActionDo {
		// 置顶
		err = infra.Dao().CommentDao.DoPin(ctx, oid, commentId)
	} else {
		// 取消置顶
		err = infra.Dao().CommentDao.SetUnPin(ctx, commentId, oid)
	}

	if err != nil {
		return xerror.Wrapf(err, "comment interact biz pin comment failed").
			WithExtras("oid", oid, "cid", commentId, "action", action).
			WithCtx(ctx)
	}

	return nil
}

func (b *CommentInteractBiz) getLikeOrDislikeCount(ctx context.Context, commentId int64, bizcode int32) (int64, error) {
	summary, err := dep.GetCounter().GetSummary(ctx, &counterv1.GetSummaryRequest{
		BizCode: bizcode,
		Oid:     commentId,
	})
	if err != nil {
		return 0, xerror.Wrapf(err, "comment interact biz counter get summary failed").WithCtx(ctx)
	}

	return summary.Count, nil
}

func (b *CommentInteractBiz) batchGetLikeOrDislikeCount(ctx context.Context, commentIds []int64, bizcode int32) (map[int64]int64, error) {
	requests := make([]*counterv1.GetSummaryRequest, 0, len(commentIds))
	for _, commentId := range commentIds {
		requests = append(requests, &counterv1.GetSummaryRequest{
			BizCode: bizcode,
			Oid:     commentId,
		})
	}

	summaries, err := dep.GetCounter().BatchGetSummary(ctx,
		&counterv1.BatchGetSummaryRequest{
			Requests: requests,
		})

	if err != nil {
		return nil, xerror.Wrapf(err, "comment interact biz counter batch get summaries failed").WithCtx(ctx)
	}

	var resp = make(map[int64]int64, len(summaries.Responses))
	for _, v := range summaries.Responses {
		resp[v.Oid] = v.Count
	}

	return resp, nil
}

// 获取评论点赞数量
func (b *CommentInteractBiz) CountCommentLikes(ctx context.Context, commentId int64) (int64, error) {
	return b.getLikeOrDislikeCount(ctx, commentId, global.CommentLikeBizcode)
}

// 批量获取评论点赞数量
func (b *CommentInteractBiz) BatchCountCommentLikes(ctx context.Context, commentIds []int64) (map[int64]int64, error) {
	return b.batchGetLikeOrDislikeCount(ctx, commentIds, global.CommentLikeBizcode)
}

// 获取评论点踩数量
func (b *CommentInteractBiz) CountCommentDislikes(ctx context.Context, commentId int64) (int64, error) {
	return b.getLikeOrDislikeCount(ctx, commentId, global.CommentDislikeBizcode)
}

// 批量获取评论点踩数量
func (b *CommentInteractBiz) BatchCountCommentDislikes(ctx context.Context, commentIds []int64) (map[int64]int64, error) {
	return b.batchGetLikeOrDislikeCount(ctx, commentIds, global.CommentDislikeBizcode)
}

// 填充点赞数量
func (b *CommentInteractBiz) PopulateLike(ctx context.Context, item *model.CommentItem) error {
	replies := []*model.CommentItem{item}
	return b.PopulateLikes(ctx, replies)
}

// 批量充评论的点赞数量
func (b *CommentInteractBiz) PopulateLikes(ctx context.Context, items []*model.CommentItem) error {
	return b.populateLikesOrHates(ctx, items, global.CommentLikeBizcode)
}

// 填充评论的点踩数量
func (b *CommentInteractBiz) PopulateHate(ctx context.Context, item *model.CommentItem) error {
	replies := []*model.CommentItem{item}
	return b.PopulateHates(ctx, replies)
}

// 批量填充评论的点踩数量
func (b *CommentInteractBiz) PopulateHates(ctx context.Context, items []*model.CommentItem) error {
	return b.populateLikesOrHates(ctx, items, global.CommentDislikeBizcode)
}

// 填充评论的点赞/点踩数量
func (b *CommentInteractBiz) populateLikesOrHates(ctx context.Context, items []*model.CommentItem, biz int32) error {
	if len(items) == 0 {
		return nil
	}

	requests := make([]*counterv1.GetSummaryRequest, 0, 16)
	for _, item := range items {
		requests = append(requests, &counterv1.GetSummaryRequest{
			BizCode: biz,
			Oid:     item.Id,
		})
	}
	resp, err := dep.GetCounter().BatchGetSummary(
		ctx, &counterv1.BatchGetSummaryRequest{
			Requests: requests,
		})
	if err != nil {
		return xerror.Wrapf(err, "comment interact biz counter batch get summary failed").
			WithExtra("len", len(items)).WithCtx(ctx)
	}

	type key struct {
		BizCode int32
		Oid     int64
	}

	mapping := make(map[key]int64, len(resp.Responses))
	for _, item := range resp.Responses {
		mapping[key{item.BizCode, item.Oid}] = item.Count
	}

	for _, cmt := range items {
		k := key{biz, cmt.Id}
		if cnt, ok := mapping[k]; ok {
			switch biz {
			case global.CommentLikeBizcode:
				cmt.LikeCount = cnt
			case global.CommentDislikeBizcode:
				cmt.HateCount = cnt
			}
		}
	}

	return nil
}

type CommentLikeStatus struct {
	CommentId int64
	Liked     bool
}

// 检查用户是否点赞过某些评论
func (b *CommentInteractBiz) BatchCheckUserLikeStatus(ctx context.Context,
	uidCommentIds map[int64][]int64) (map[int64][]CommentLikeStatus, error) {

	params := make(map[int64]*counterv1.ObjectList)
	for uid, commentIds := range uidCommentIds {
		params[uid] = &counterv1.ObjectList{
			Oids: commentIds,
		}
	}
	resp, err := dep.GetCounter().BatchGetRecord(ctx,
		&counterv1.BatchGetRecordRequest{
			BizCode: global.CommentLikeBizcode,
			Params:  params,
		})
	if err != nil {
		return nil, xerror.Wrapf(err, "comment interact biz counter batch get record failed").WithCtx(ctx)
	}

	var result = make(map[int64][]CommentLikeStatus, len(resp.GetResults()))
	for uid, commentIds := range uidCommentIds {
		likeRecords := resp.GetResults()[uid]
		for _, commentId := range commentIds {
			liked := false
			for _, likeRecord := range likeRecords.GetList() {
				if likeRecord.Oid == commentId && likeRecord.Act == counterv1.RecordAct_RECORD_ACT_ADD {
					liked = true
				}
			}

			result[uid] = append(result[uid], CommentLikeStatus{
				CommentId: commentId,
				Liked:     liked,
			})
		}
	}

	return result, nil
}
