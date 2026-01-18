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
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/comment/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	globalmodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/errors"
)

type Biz struct{}

func NewBiz() *Biz { return &Biz{} }

// PublishNoteComment 发布评论
func (b *Biz) PublishNoteComment(ctx context.Context, req *commentv1.AddCommentRequest) (int64, error) {
	if err := b.checkPubReqValid(ctx, req); err != nil {
		return 0, err
	}

	resp, err := dep.Commenter().AddComment(ctx, req)
	if err != nil {
		return 0, xerror.Wrapf(err, "remote commenter add comment failed")
	}

	return resp.CommentId, nil
}

func (b *Biz) checkPubReqValid(ctx context.Context, req *commentv1.AddCommentRequest) error {
	pubOnOidDirectly := req.RootId == 0 && req.ParentId == 0
	if pubOnOidDirectly {
		// comment on note directly
		resp, err := dep.NoteFeedServer().GetNoteAuthor(ctx, &notev1.GetNoteAuthorRequest{
			NoteId: req.Oid,
		})
		if err != nil {
			return xerror.Wrapf(err, "remote check note author when pub on note failed").
				WithExtras("req", req).WithCtx(ctx)
		}

		if resp.Author != req.ReplyUid {
			return xerror.Wrap(errors.ErrReplyUserDoesNotMatch)
		}
	} else {
		// comment on comment, check parent comment
		resp, err := dep.Commenter().GetCommentUser(ctx, &commentv1.GetCommentUserRequest{
			CommentId: req.ParentId,
		})
		if err != nil {
			return xerror.Wrapf(err, "remote check parent comment when pub on note failed").
				WithExtras("req", req).WithCtx(ctx)
		}
		if resp.Uid != req.ReplyUid {
			return xerror.Wrap(errors.ErrReplyUserDoesNotMatch)
		}
	}

	return nil
}

// PageGetNoteRootComments 分页获取主评论
func (b *Biz) PageGetNoteRootComments(ctx context.Context, req *commentv1.PageGetCommentRequest) (*model.CommentRes, error) {
	rootReplies, err := dep.Commenter().PageGetComment(ctx, req)
	if err != nil {
		return nil, xerror.Wrapf(err, "remote commenter page get comment failed")
	}

	// 整合用户的信息
	comments := []*model.CommentItem{}
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

// PageGetNoteSubComments 分页获取子评论
func (b *Biz) PageGetNoteSubComments(ctx context.Context, req *commentv1.PageGetSubCommentRequest) (*model.CommentRes, error) {
	subReplies, err := dep.Commenter().PageGetSubComment(ctx, req)
	if err != nil {
		return nil, xerror.Wrapf(err, "remote commenter page get sub comment failed")
	}

	// 填充评论的用户信息
	comments := []*model.CommentItem{}
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

func (b *Biz) getPinnedComment(ctx context.Context, oid int64) (*model.DetailedCommentItem, error) {
	var comment *commentv1.DetailedCommentItem

	resp, err := dep.Commenter().GetPinnedComment(ctx,
		&commentv1.GetPinnedCommentRequest{Oid: oid})
	if err != nil {
		return nil, xerror.Wrapf(err, "remote get pinned comment failed").WithCtx(ctx)
	}
	comment = resp.GetItem()
	if comment.GetRoot() == nil {
		// 可能不存在置顶评论
		return nil, nil
	}

	if comment != nil && comment.GetRoot() != nil {
		userResp, err := dep.Userer().
			BatchGetUser(ctx,
				&userv1.BatchGetUserRequest{
					Uids: xmap.Keys(extractUidsMap([]*commentv1.DetailedCommentItem{comment})),
				},
			)
		if err != nil {
			return nil, xerror.Wrapf(err, "remote batch get user failed").WithCtx(ctx)
		}

		return model.NewDetailedCommentItemFromPb(comment, userResp.Users), nil
	}

	return nil, nil // before careful
}

func (b *Biz) getSeekedComment(ctx context.Context, oid, seekId int64) (*model.DetailedCommentItem, error) {
	var seekedComment *commentv1.DetailedCommentItem

	resp, err := dep.Commenter().GetComment(ctx,
		&commentv1.GetCommentRequest{
			CommentId: seekId,
		})
	if err != nil {
		return nil, xerror.Wrapf(err, "remote get seek comment failed").WithCtx(ctx)
	}

	seekCommentResp := resp.GetItem()
	if seekCommentResp.Oid != oid { // oid does not match
		return nil, nil
	}

	seekedComment = &commentv1.DetailedCommentItem{}
	if seekCommentResp.RootId == 0 && seekCommentResp.ParentId == 0 {
		// seekComment is root, we need to fetch sub comments
		seekSubComments, err := dep.Commenter().PageGetSubComment(ctx,
			&commentv1.PageGetSubCommentRequest{
				Oid:    seekCommentResp.Oid,
				RootId: seekCommentResp.Id,
				Cursor: 0,
			})
		if err != nil {
			return nil, xerror.Wrapf(err, "remote get page sub comment failed").
				WithExtras("oid", seekCommentResp.Oid, "root_id", seekCommentResp.RootId, "seek_id", seekId).WithCtx(ctx)
		}

		seekedComment.Root = seekCommentResp
		seekedComment.SubComments = &commentv1.DetailedSubComment{
			Items:      seekSubComments.Comments,
			HasNext:    seekSubComments.HasNext,
			NextCursor: seekSubComments.NextCursor,
		}
	} else {
		// seekComment is not root, we need to fetch root comment
		seekRootComment, err := dep.Commenter().GetComment(ctx,
			&commentv1.GetCommentRequest{
				CommentId: seekCommentResp.RootId,
			})
		if err != nil {
			return nil, xerror.Wrapf(err, "remote get comment failed").
				WithExtras("oid", seekCommentResp.Oid, "root_id", seekCommentResp.RootId, "seek_id", seekId).WithCtx(ctx)
		}
		seekedComment.Root = seekRootComment.Item
		seekedComment.SubComments = &commentv1.DetailedSubComment{
			Items:      []*commentv1.CommentItem{seekCommentResp},
			NextCursor: 0,
			HasNext:    true,
		}
	}

	if seekedComment.GetRoot() != nil {
		// extra user
		userResp, err := dep.Userer().
			BatchGetUser(ctx,
				&userv1.BatchGetUserRequest{
					Uids: xmap.Keys(extractUidsMap([]*commentv1.DetailedCommentItem{seekedComment})),
				},
			)
		if err != nil {
			return nil, xerror.Wrapf(err, "remote batch get user failed").WithCtx(ctx)
		}

		return model.NewDetailedCommentItemFromPb(seekedComment, userResp.Users), nil
	}

	return nil, nil // be careful
}

// PageGetNoteCommentsReq 获取评论请求参数
type PageGetNoteCommentsReq struct {
	Oid    int64
	Cursor int64
	SortBy int32
	SeekId int64
}

// PageGetNoteComments 获取主评论信息（包含其下子评论）
func (b *Biz) PageGetNoteComments(ctx context.Context, req *PageGetNoteCommentsReq) (*model.DetailedCommentRes, error) {
	var (
		wg sync.WaitGroup

		pinned *model.DetailedCommentItem // 笔记置顶评论
		seeked *model.DetailedCommentItem // 特定需要的评论 如果存在会放在返回值开头
	)

	if req.Cursor == 0 {
		wg.Add(1)
		// 第一次请求时需要返回置顶评论
		concurrent.SafeGo(func() {
			defer wg.Done()
			var err error
			pinned, err = b.getPinnedComment(ctx, req.Oid)
			if err != nil {
				xlog.Msg("get pinned comment failed").Extras("req", req).Err(err).Errorx(ctx)
			}
		})
	}

	// check if need seek comment
	if req.SeekId != 0 {
		wg.Add(1)
		concurrent.SafeGo(func() {
			defer wg.Done()
			var err error
			seeked, err = b.getSeekedComment(ctx, req.Oid, req.SeekId)
			if err != nil {
				xlog.Msg("get seeked comment failed").Extras("req", req).Err(err).Errorx(ctx)
			}
		})
	}

	resp, err := dep.Commenter().PageGetDetailedComment(ctx, &commentv1.PageGetDetailedCommentRequest{
		Oid:    req.Oid,
		Cursor: req.Cursor,
		SortBy: commentv1.SortType(req.SortBy),
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "remote commenter page get detailed comment failed")
	}

	comments := []*model.DetailedCommentItem{}

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

	wg.Wait()

	temps := make([]*model.DetailedCommentItem, 0, len(comments)+2)
	temps = append(temps, comments...)
	if pinned != nil {
		temps = append(temps, pinned)
	}
	if req.SeekId != 0 && seeked != nil {
		temps = append(temps, seeked)
	}

	attachDetailCommentItemInteract(ctx, temps)

	// prepend
	if req.SeekId != 0 && seeked != nil {
		newComments := make([]*model.DetailedCommentItem, 0, len(comments)+1)
		newComments = append(newComments, seeked)
		newComments = append(newComments, comments...)
		comments = xslice.UniqF(newComments, func(v *model.DetailedCommentItem) int64 { return v.Root.Id })
	}

	return &model.DetailedCommentRes{
		Comments:   comments,
		PinComment: pinned,
		HasNext:    resp.HasNext,
		NextCursor: resp.NextCursor,
	}, nil
}

// DelNoteComment 删除评论
func (b *Biz) DelNoteComment(ctx context.Context, commentId, oid int64) error {
	_, err := dep.Commenter().DelComment(ctx, &commentv1.DelCommentRequest{
		CommentId: commentId,
		Oid:       oid,
	})
	if err != nil {
		return xerror.Wrapf(err, "remote commenter del comment failed")
	}

	return err
}

// PinNoteComment 置顶评论
func (b *Biz) PinNoteComment(ctx context.Context, oid, commentId int64, action int8) error {
	_, err := dep.Commenter().PinComment(ctx, &commentv1.PinCommentRequest{
		Oid:       oid,
		CommentId: commentId,
		Action:    commentv1.CommentAction(action),
	})
	if err != nil {
		return xerror.Wrapf(err, "remote commenter pin comment failed")
	}

	return err
}

// LikeNoteComment 点赞评论
func (b *Biz) LikeNoteComment(ctx context.Context, commentId int64, action uint8) error {
	_, err := dep.Commenter().LikeAction(ctx, &commentv1.LikeActionRequest{
		CommentId: commentId,
		Action:    commentv1.CommentAction(action),
	})
	if err != nil {
		return xerror.Wrapf(err, "remote commenter like action failed")
	}

	return err
}

// DislikeNoteComment 点踩评论
func (b *Biz) DislikeNoteComment(ctx context.Context, commentId int64, action uint8) error {
	_, err := dep.Commenter().DislikeAction(ctx, &commentv1.DislikeActionRequest{
		CommentId: commentId,
		Action:    commentv1.CommentAction(action),
	})
	if err != nil {
		return xerror.Wrapf(err, "remote commenter dislike action failed")
	}

	return err
}

// GetNoteCommentLikeCount 获取评论点赞数
func (b *Biz) GetNoteCommentLikeCount(ctx context.Context, commentId int64) (int64, error) {
	resp, err := dep.Commenter().GetCommentLikeCount(ctx, &commentv1.GetCommentLikeCountRequest{
		CommentId: commentId,
	})
	if err != nil {
		return 0, xerror.Wrapf(err, "remote commenter get comment like count failed")
	}

	return resp.Count, nil
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

	uid := metadata.Uid(ctx)

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

func (b *Biz) GetCommentContent(ctx context.Context, id int64) (*globalmodel.CommentContent, error) {
	resp, err := dep.Commenter().GetComment(ctx, &commentv1.GetCommentRequest{
		CommentId: id,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "remote get comment by id failed").WithExtra("comment_id", id).WithCtx(ctx)
	}

	content := resp.Item.Content

	atUsers := make([]globalmodel.AtUser, 0, len(resp.Item.AtUsers))
	for _, au := range resp.Item.AtUsers {
		atUsers = append(atUsers, globalmodel.AtUser{
			Uid:      au.Uid,
			Nickname: au.Nickname,
		})
	}

	return &globalmodel.CommentContent{
		Text:    content,
		AtUsers: atUsers,
	}, nil
}

func (b *Biz) GetCommentUser(ctx context.Context, id int64) (int64, error) {
	resp, err := dep.Commenter().GetCommentUser(ctx, &commentv1.GetCommentUserRequest{
		CommentId: id,
	})
	if err != nil {
		return 0, xerror.Wrapf(err, "remote get comment user failed").WithExtra("comment_id", id).WithCtx(ctx)
	}

	return resp.GetUid(), nil
}

func (b *Biz) GetComment(ctx context.Context, id int64) (*model.CommentItemBase, error) {
	resp, err := dep.Commenter().GetComment(ctx, &commentv1.GetCommentRequest{
		CommentId: id,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "remote get comment failed").WithExtra("comment_id", id).WithCtx(ctx)
	}

	return model.NewCommentItemBaseFromPb(resp.GetItem()), nil
}
