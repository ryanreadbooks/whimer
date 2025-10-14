package comment

import (
	"context"
	"strconv"
	"sync"

	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/comment/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	globalmodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
)

type Biz struct {
}

func NewBiz() *Biz { return &Biz{} }

func (b *Biz) PublishNoteComment(ctx context.Context, req *model.PubReq) (*model.PubRes, error) {
	resp, err := dep.Commenter().AddComment(ctx, req.AsPb())
	if err != nil {
		return nil, xerror.Wrapf(err, "remote commenter add comment failed")
	}

	return &model.PubRes{CommentId: resp.CommentId}, nil
}

func (b *Biz) PageGetNoteRootComments(ctx context.Context, req *model.GetCommentsReq) (*model.CommentRes, error) {
	rootReplies, err := dep.Commenter().PageGetComment(ctx, req.AsPb())
	if err != nil {
		return nil, xerror.Wrapf(err, "remote commenter page get comment failed")
	}

	// 整合用户的信息
	var comments = []*model.CommentItem{}
	if len(rootReplies.Comments) > 0 {
		comments, err = attachCommentsUsers(ctx, rootReplies.Comments)
		if err != nil {
			return nil, xerror.Wrapf(err, "attach comments user failed")
		}
	}

	attachCommentItemInteract(ctx, comments)

	return &model.CommentRes{
		Items:      comments,
		NextCursor: rootReplies.NextCursor,
		HasNext:    rootReplies.HasNext,
	}, nil
}

func (b *Biz) PageGetNoteSubComments(ctx context.Context, req *model.GetSubCommentsReq) (*model.CommentRes, error) {
	subReplies, err := dep.Commenter().PageGetSubComment(ctx, req.AsPb())
	if err != nil {
		return nil, xerror.Wrapf(err, "remote commenter page get sub comment failed")
	}

	// 填充评论的用户信息
	var comments = []*model.CommentItem{}
	if len(subReplies.Comments) > 0 {
		comments, err = attachCommentsUsers(ctx, subReplies.Comments)
		if err != nil {
			return nil, xerror.Wrapf(err, "attach comments user failed")
		}
	}

	attachCommentItemInteract(ctx, comments)

	return &model.CommentRes{
		Items:      comments,
		NextCursor: subReplies.NextCursor,
		HasNext:    subReplies.HasNext,
	}, nil
}

func (b *Biz) PageGetNoteComments(ctx context.Context, req *model.GetCommentsReq) (*model.DetailedCommentRes, error) {
	var (
		pinnedReply     *commentv1.DetailedCommentItem
		pinnedReplyUser map[string]*userv1.UserInfo = nil
		wg              sync.WaitGroup
	)

	if req.Cursor == 0 {
		wg.Add(1)
		// 第一次请求时需要返回置顶评论
		concurrent.SafeGo(func() {
			defer wg.Done()
			resp, err := dep.Commenter().GetPinnedComment(ctx, &commentv1.GetPinnedCommentRequest{Oid: int64(req.Oid)})
			if err != nil {
				xlog.Msg("get pinned comment failed").Err(err).Errorx(ctx)
				return
			}
			pinnedReply = resp.GetItem()
			if pinnedReply.GetRoot() == nil {
				// 可能不存在置顶评论
				return
			}

			userResp, err := dep.Userer().
				BatchGetUser(ctx,
					&userv1.BatchGetUserRequest{
						Uids: xmap.Keys(extractUidsMap([]*commentv1.DetailedCommentItem{pinnedReply})),
					},
				)
			if err != nil {
				xlog.Msg("get batch get user failed").Err(err).Errorx(ctx)
				return
			}
			pinnedReplyUser = make(map[string]*userv1.UserInfo)
			pinnedReplyUser = userResp.Users
		})
	}

	resp, err := dep.Commenter().PageGetDetailedComment(ctx, req.AsDetailedPb())
	if err != nil {
		return nil, xerror.Wrapf(err, "remote commenter page get detailed comment failed")
	}

	var (
		comments = []*model.DetailedCommentItem{}
	)

	if len(resp.Comments) > 0 {
		uidsMap := extractUidsMap(resp.Comments)

		// 发起请求获取uid的详细信息
		userResp, err := dep.Userer().BatchGetUser(ctx,
			&userv1.BatchGetUserRequest{Uids: xmap.Keys(uidsMap)})
		if err != nil {
			return nil, xerror.Wrapf(err, "remote userer batch get user failed")
		}

		// 拼接结果
		comments = make([]*model.DetailedCommentItem, 0, len(resp.Comments))
		for _, item := range resp.Comments {
			details := model.NewDetailedCommentItemFromPb(item, userResp.Users)
			comments = append(comments, details)
		}
	}

	var pinned *model.DetailedCommentItem
	wg.Wait()
	if req.Cursor == 0 {
		// 置顶
		if pinnedReply != nil && pinnedReply.GetRoot() != nil { // 有些可能没有设置置顶评论
			pinned = model.NewDetailedCommentItemFromPb(pinnedReply, pinnedReplyUser)
		}
	}

	temps := make([]*model.DetailedCommentItem, 0, len(comments)+1)
	temps = append(temps, comments...)
	if pinned != nil {
		temps = append(temps, pinned)
	}
	attachDetailCommentItemInteract(ctx, temps)
	return &model.DetailedCommentRes{
		Comments:   comments,
		PinComment: pinned,
		HasNext:    resp.HasNext,
		NextCursor: resp.NextCursor,
	}, nil
}

func (b *Biz) DelNoteComment(ctx context.Context, req *model.DelReq) error {
	_, err := dep.Commenter().DelComment(ctx, &commentv1.DelCommentRequest{
		CommentId: req.CommentId,
		Oid:       int64(req.Oid),
	})
	if err != nil {
		return xerror.Wrapf(err, "remote commenter del comment failed")
	}

	return err
}

func (b *Biz) PinNoteComment(ctx context.Context, req *model.PinReq) error {
	_, err := dep.Commenter().PinComment(ctx, &commentv1.PinCommentRequest{
		Oid:       int64(req.Oid),
		CommentId: req.CommentId,
		Action:    commentv1.CommentAction(req.Action),
	})
	if err != nil {
		return xerror.Wrapf(err, "remote commenter pin comment failed")
	}

	return err
}

func (b *Biz) LikeNoteComment(ctx context.Context, req *model.ThumbUpReq) error {
	_, err := dep.Commenter().LikeAction(ctx, &commentv1.LikeActionRequest{
		CommentId: req.CommentId,
		Action:    commentv1.CommentAction(req.Action),
	})
	if err != nil {
		return xerror.Wrapf(err, "remote commenter like action failed")
	}

	return err
}

func (b *Biz) DislikeNoteComment(ctx context.Context, req *model.ThumbDownReq) error {
	_, err := dep.Commenter().DislikeAction(ctx, &commentv1.DislikeActionRequest{
		CommentId: req.CommentId,
		Action:    commentv1.CommentAction(req.Action),
	})
	if err != nil {
		return xerror.Wrapf(err, "remote commenter dislike action failed")
	}

	return err
}

func (b *Biz) GetNoteCommentLikeCount(ctx context.Context, req *model.GetLikeCountReq) (*model.GetLikeCountRes, error) {
	resp, err := dep.Commenter().GetCommentLikeCount(ctx, &commentv1.GetCommentLikeCountRequest{
		CommentId: req.CommentId,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "remote commenter get comment like count failed")
	}

	return &model.GetLikeCountRes{
		Comment: resp.CommentId,
		Likes:   resp.Count,
	}, nil
}

func (b *Biz) UploadCommentImages(ctx context.Context, req *model.UploadImagesReq) (*model.UploadImagesRes, error) {
	resp, err := dep.Commenter().UploadCommentImages(ctx, &commentv1.UploadCommentImagesRequest{
		RequestedCount: req.Count,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "remote commenter upload comment images failed")
	}

	return resp, nil
}

// 填入用户信息
func attachCommentsUsers(ctx context.Context, comments []*commentv1.CommentItem) ([]*model.CommentItem, error) {
	uidsMap := make(map[int64]struct{})
	for _, root := range comments {
		uidsMap[root.Uid] = struct{}{}
	}

	userResp, err := dep.Userer().
		BatchGetUser(ctx, &userv1.BatchGetUserRequest{Uids: xmap.Keys(uidsMap)})
	if err != nil {
		return nil, xerror.Wrapf(err, "remote userer batch get user failed")
	}

	res := make([]*model.CommentItem, 0, len(comments))
	for _, root := range comments {
		res = append(res, &model.CommentItem{
			CommentItemBase: model.NewCommentItemBaseFromPb(root),
			User:            userResp.Users[formatUid(root.Uid)],
		})
	}

	return res, nil
}

func formatUid(uid int64) string {
	return strconv.FormatInt(uid, 10)
}

func attachCommentItemInteract(ctx context.Context, items []*model.CommentItem) {
	if globalmodel.IsGuestFromCtx(ctx) {
		return
	}

	var uid = metadata.Uid(ctx)

	if len(items) == 0 {
		return
	}

	// collect all comment ids
	commentIds := make([]int64, 0, len(items))
	for _, item := range items {
		commentIds = append(commentIds, item.Id)
	}

	commentIds = xslice.Uniq(commentIds)

	// BatchCheckUserLikeReply有数量限制 此处需要分批处理
	var wg sync.WaitGroup
	var syncLikeStatus sync.Map
	err := xslice.BatchAsyncExec(&wg, commentIds, 50, func(start, end int) error {
		resp, err := dep.Commenter().BatchCheckUserLikeComment(ctx,
			&commentv1.BatchCheckUserLikeCommentRequest{
				Mappings: map[int64]*commentv1.BatchCheckUserLikeCommentRequest_CommentIdList{
					uid: {Ids: commentIds[start:end]},
				},
			})
		if err != nil {
			return err
		}

		if status, ok := resp.GetResults()[uid]; ok {
			for _, status := range status.List {
				syncLikeStatus.Store(status.GetCommentId(), status.GetLiked())
			}
		}

		return nil
	})

	if err != nil {
		xlog.Msg("comment biz failed to check user like comment status").Errorx(ctx)
		return
	}

	// fill items
	for _, item := range items {
		if v, ok := syncLikeStatus.Load(item.Id); ok {
			if vv, yes := v.(bool); yes {
				item.Interact.Liked = vv
			}
		}
	}
}

func attachDetailCommentItemInteract(ctx context.Context, dItems []*model.DetailedCommentItem) {
	items := make([]*model.CommentItem, 0, len(dItems))
	for _, dItem := range dItems {
		items = append(items, dItem.Root)
		items = append(items, dItem.SubComments.Items...)
	}

	attachCommentItemInteract(ctx, items)
}

func extractUidsMap(replies []*commentv1.DetailedCommentItem) map[int64]struct{} {
	uidsMap := make(map[int64]struct{})
	// 提取出主评论和子评论的uid
	for _, item := range replies {
		uidsMap[item.Root.Uid] = struct{}{}
		for _, sub := range item.SubComments.Items {
			uidsMap[sub.Uid] = struct{}{}
		}
	}

	return uidsMap
}
