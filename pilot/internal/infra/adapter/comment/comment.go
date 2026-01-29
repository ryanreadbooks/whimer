package comment

import (
	"context"
	"sync"

	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/comment/vo"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/comment/convert"
	notecache "github.com/ryanreadbooks/whimer/pilot/internal/infra/core/cache/note"
)

type CommentAdapterImpl struct {
	commentCli    commentv1.CommentServiceClient
	noteStatCache *notecache.StatStore
}

func NewCommentAdapterImpl(
	commentCli commentv1.CommentServiceClient,
	noteStatCache *notecache.StatStore,
) *CommentAdapterImpl {
	return &CommentAdapterImpl{
		commentCli:    commentCli,
		noteStatCache: noteStatCache,
	}
}

var _ repository.CommentAdapter = (*CommentAdapterImpl)(nil)

func (a *CommentAdapterImpl) BatchCheckCommented(ctx context.Context,
	p *repository.BatchCheckCommentedParams,
) (*repository.BatchCheckCommentedResult, error) {
	m := make(map[int64]*commentv1.BatchCheckUserOnObjectRequest_Objects)
	m[p.Uid] = &commentv1.BatchCheckUserOnObjectRequest_Objects{Oids: p.NoteIds}
	req := &commentv1.BatchCheckUserOnObjectRequest{
		Mappings: m,
	}
	resp, err := a.commentCli.BatchCheckUserOnObject(ctx, req)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	pairs := resp.GetResults()
	commented := make(map[int64]bool, len(pairs))
	for _, item := range pairs[p.Uid].GetList() {
		commented[item.GetOid()] = item.GetCommented()
	}

	return &repository.BatchCheckCommentedResult{
		Commented: commented,
	}, nil
}

func (a *CommentAdapterImpl) BatchCheckCommentExist(ctx context.Context, commentIds []int64) (map[int64]bool, error) {
	resp, err := a.commentCli.BatchCheckCommentExist(ctx, &commentv1.BatchCheckCommentExistRequest{
		Ids: commentIds,
	})
	if err != nil {
		return nil, xerror.Wrap(err).WithCtx(ctx)
	}
	return resp.GetExistence(), nil
}

func (a *CommentAdapterImpl) BatchCheckUsersLikeComment(ctx context.Context,
	mappings map[int64][]int64,
) (map[int64]map[int64]bool, error) {
	reqMappings := make(map[int64]*commentv1.BatchCheckUserLikeCommentRequest_CommentIdList)
	for uid, commentIds := range mappings {
		reqMappings[uid] = &commentv1.BatchCheckUserLikeCommentRequest_CommentIdList{Ids: commentIds}
	}

	resp, err := a.commentCli.BatchCheckUserLikeComment(ctx, &commentv1.BatchCheckUserLikeCommentRequest{
		Mappings: reqMappings,
	})
	if err != nil {
		return nil, xerror.Wrap(err).WithCtx(ctx)
	}

	result := make(map[int64]map[int64]bool)
	for uid, statusList := range resp.GetResults() {
		result[uid] = make(map[int64]bool)
		for _, item := range statusList.GetList() {
			result[uid][item.GetCommentId()] = item.GetLiked()
		}
	}

	return result, nil
}

func (a *CommentAdapterImpl) AddComment(ctx context.Context, p *repository.AddCommentParams) (int64, error) {
	images := make([]*commentv1.CommentReqImage, 0, len(p.Images))
	for _, img := range p.Images {
		images = append(images, &commentv1.CommentReqImage{
			StoreKey: img.StoreKey,
			Width:    img.Width,
			Height:   img.Height,
			Format:   img.Format,
		})
	}

	atUsers := make([]*commentv1.CommentAtUser, 0, len(p.AtUsers))
	for _, au := range p.AtUsers {
		atUsers = append(atUsers, &commentv1.CommentAtUser{
			Uid:      au.Uid,
			Nickname: au.Nickname,
		})
	}

	resp, err := a.commentCli.AddComment(ctx,
		&commentv1.AddCommentRequest{
			Type:     commentv1.CommentType(p.Type),
			Oid:      p.Oid,
			Content:  p.Content,
			RootId:   p.RootId,
			ParentId: p.ParentId,
			ReplyUid: p.ReplyUid,
			Images:   images,
			AtUsers:  atUsers,
		})
	if err != nil {
		return 0, xerror.Wrap(err)
	}

	// 同步评论增量到缓存
	a.syncCommentCount(ctx, p.Oid, 1)

	return resp.CommentId, nil
}

func (a *CommentAdapterImpl) GetComment(ctx context.Context, commentId int64) (*entity.Comment, error) {
	resp, err := a.commentCli.GetComment(ctx,
		&commentv1.GetCommentRequest{
			CommentId: commentId,
		})
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return convert.PbCommentItemToEntity(resp.GetItem()), nil
}

func (a *CommentAdapterImpl) GetCommentUser(ctx context.Context, commentId int64) (int64, error) {
	resp, err := a.commentCli.GetCommentUser(
		ctx,
		&commentv1.GetCommentUserRequest{
			CommentId: commentId,
		})
	if err != nil {
		return 0, xerror.Wrap(err)
	}

	return resp.GetUid(), nil
}

func (a *CommentAdapterImpl) DelComment(ctx context.Context, commentId, oid int64) error {
	_, err := a.commentCli.DelComment(ctx,
		&commentv1.DelCommentRequest{
			CommentId: commentId,
			Oid:       oid,
		})
	if err != nil {
		return xerror.Wrap(err)
	}

	// 同步评论减少
	a.syncCommentCount(ctx, oid, -1)

	return nil
}

func (a *CommentAdapterImpl) PinComment(ctx context.Context, oid, commentId int64, action vo.PinAction) error {
	_, err := a.commentCli.PinComment(ctx,
		&commentv1.PinCommentRequest{
			Oid:       oid,
			CommentId: commentId,
			Action:    convert.PinActionAsPb(action),
		})
	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}

func (a *CommentAdapterImpl) LikeComment(ctx context.Context, commentId int64, action vo.ThumbAction) error {
	_, err := a.commentCli.LikeAction(ctx,
		&commentv1.LikeActionRequest{
			CommentId: commentId,
			Action:    convert.ThumbActionAsPb(action),
		})
	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}

func (a *CommentAdapterImpl) DislikeComment(ctx context.Context, commentId int64, action vo.ThumbAction) error {
	_, err := a.commentCli.DislikeAction(ctx,
		&commentv1.DislikeActionRequest{
			CommentId: commentId,
			Action:    convert.ThumbActionAsPb(action),
		})
	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}

func (a *CommentAdapterImpl) GetCommentLikeCount(ctx context.Context, commentId int64) (int64, error) {
	resp, err := a.commentCli.GetCommentLikeCount(ctx,
		&commentv1.GetCommentLikeCountRequest{
			CommentId: commentId,
		})
	if err != nil {
		return 0, xerror.Wrap(err)
	}

	return resp.Count, nil
}

func (a *CommentAdapterImpl) BatchCheckUserLikeComment(
	ctx context.Context, uid int64, commentIds []int64,
) (map[int64]bool, error) {
	var wg sync.WaitGroup
	var syncLikeStatus sync.Map

	err := xslice.BatchAsyncExec(&wg, commentIds, 50, func(start, end int) error {
		resp, err := a.commentCli.BatchCheckUserLikeComment(ctx,
			&commentv1.BatchCheckUserLikeCommentRequest{
				Mappings: map[int64]*commentv1.BatchCheckUserLikeCommentRequest_CommentIdList{
					uid: {Ids: commentIds[start:end]},
				},
			})
		if err != nil {
			return err
		}

		if status, ok := resp.GetResults()[uid]; ok {
			for _, s := range status.List {
				syncLikeStatus.Store(s.GetCommentId(), s.GetLiked())
			}
		}

		return nil
	})
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	result := make(map[int64]bool, len(commentIds))
	syncLikeStatus.Range(func(key, value any) bool {
		if id, ok := key.(int64); ok {
			if liked, ok := value.(bool); ok {
				result[id] = liked
			}
		}
		return true
	})

	return result, nil
}

func (a *CommentAdapterImpl) PageGetRootComments(
	ctx context.Context, p *repository.PageGetCommentsParams,
) (*repository.PageGetCommentsResult, error) {
	resp, err := a.commentCli.PageGetComment(ctx,
		&commentv1.PageGetCommentRequest{
			Oid:    p.Oid,
			Cursor: p.Cursor,
			SortBy: commentv1.SortType(p.SortBy),
		})
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	items := make([]*entity.Comment, 0, len(resp.Comments))
	for _, item := range resp.Comments {
		items = append(items, convert.PbCommentItemToEntity(item))
	}

	return &repository.PageGetCommentsResult{
		Items:      items,
		NextCursor: resp.NextCursor,
		HasNext:    resp.HasNext,
	}, nil
}

func (a *CommentAdapterImpl) PageGetSubComments(
	ctx context.Context, p *repository.PageGetSubCommentsParams,
) (*repository.PageGetCommentsResult, error) {
	resp, err := a.commentCli.PageGetSubComment(ctx,
		&commentv1.PageGetSubCommentRequest{
			Oid:    p.Oid,
			RootId: p.RootId,
			Cursor: p.Cursor,
		})
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	items := make([]*entity.Comment, 0, len(resp.Comments))
	for _, item := range resp.Comments {
		items = append(items, convert.PbCommentItemToEntity(item))
	}

	return &repository.PageGetCommentsResult{
		Items:      items,
		NextCursor: resp.NextCursor,
		HasNext:    resp.HasNext,
	}, nil
}

func (a *CommentAdapterImpl) PageGetDetailedComments(
	ctx context.Context, p *repository.PageGetDetailedCommentsParams,
) (*repository.PageGetDetailedCommentsResult, error) {
	resp, err := a.commentCli.PageGetDetailedComment(ctx,
		&commentv1.PageGetDetailedCommentRequest{
			Oid:    p.Oid,
			Cursor: p.Cursor,
			SortBy: commentv1.SortType(p.SortBy),
		})
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	items := make([]*entity.DetailedComment, 0, len(resp.Comments))
	for _, item := range resp.Comments {
		items = append(items, convert.PbDetailedCommentItemToEntity(item))
	}

	return &repository.PageGetDetailedCommentsResult{
		Items:      items,
		NextCursor: resp.NextCursor,
		HasNext:    resp.HasNext,
	}, nil
}

func (a *CommentAdapterImpl) GetPinnedComment(ctx context.Context, oid int64) (*entity.DetailedComment, error) {
	resp, err := a.commentCli.GetPinnedComment(ctx,
		&commentv1.GetPinnedCommentRequest{
			Oid: oid,
		})
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	item := resp.GetItem()
	if item.GetRoot() == nil {
		return nil, nil
	}

	return convert.PbDetailedCommentItemToEntity(item), nil
}

// 同步评论数量增量到缓存
func (a *CommentAdapterImpl) syncCommentCount(ctx context.Context, oid int64, inc int64) {
	if err := a.noteStatCache.Add(
		ctx,
		notecache.NoteCommentCountStat,
		notecache.NoteStatRepr{
			NoteId: notevo.NoteId(oid).String(),
			Inc:    inc,
		}); err != nil {
		xlog.Msg("note stat add comment count failed").Err(err).
			Extras("oid", oid, "inc", inc).Errorx(ctx)
	}
}
