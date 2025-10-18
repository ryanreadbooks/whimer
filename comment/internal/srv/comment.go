package srv

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/comment/internal/biz"
	"github.com/ryanreadbooks/whimer/comment/internal/global"
	"github.com/ryanreadbooks/whimer/comment/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/comment/internal/model"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"

	"golang.org/x/sync/errgroup"
)

type CommentSrv struct {
	CommentBiz         biz.CommentBiz
	CommentInteractBiz biz.CommentInteractBiz
	AssetManagerBiz    *biz.AssetManagerBiz
}

func NewCommentSrv(s *Service, biz biz.Biz) *CommentSrv {
	return &CommentSrv{
		CommentBiz:         biz.CommentBiz,
		CommentInteractBiz: biz.CommentInteractBiz,
		AssetManagerBiz:    biz.AssetManagerBiz,
	}
}

// 用户发表评论
func (s *CommentSrv) AddComment(ctx context.Context, req *model.AddCommentReq) (*model.AddCommentRes, error) {
	res, err := s.CommentBiz.AddComment(ctx, req)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to add comment").WithCtx(ctx).WithExtra("req", req)
	}

	return res, nil
}

// 用户删除评论
func (s *CommentSrv) DelComment(ctx context.Context, oid, commentId int64) error {
	err := s.CommentBiz.DelComment(ctx, oid, commentId)
	if err != nil {
		return xerror.Wrapf(err, "comment srv failed to del comment").
			WithCtx(ctx).
			WithExtra("commentId", commentId)
	}

	return nil
}

// 用户点赞/取消点赞某条评论
func (s *CommentSrv) LikeComment(ctx context.Context, commentId int64, action int8) error {
	_, err := s.CommentBiz.GetComment(ctx, commentId,
		biz.DoNotPopulateExt(), biz.DoNotPopulateImages())
	if err != nil {
		return xerror.Wrapf(err, "comment srv pin comment failed").WithCtx(ctx)
	}

	err = s.CommentInteractBiz.LikeComment(ctx, commentId, action)
	if err != nil {
		return xerror.Wrapf(err, "comment srv failed to do like comment").
			WithCtx(ctx).
			WithExtras("commentId", commentId, "action", action)
	}

	return nil
}

// 用户点踩/取消点踩某条评论
func (s *CommentSrv) DislikeComment(ctx context.Context, commentId int64, action int8) error {
	_, err := s.CommentBiz.GetComment(ctx, commentId,
		biz.DoNotPopulateExt(), biz.DoNotPopulateImages())
	if err != nil {
		return xerror.Wrapf(err, "comment srv pin comment failed").WithCtx(ctx)
	}

	err = s.CommentInteractBiz.DislikeComment(ctx, commentId, action)
	if err != nil {
		return xerror.Wrapf(err, "comment srv failed to do dislike comment").
			WithCtx(ctx).
			WithExtras("commentId", commentId, "action", action)
	}

	return nil
}

// 置顶评论/取消置顶评论
func (s *CommentSrv) PinComment(ctx context.Context, oid, commentId int64, action int8) error {
	var (
		uid = metadata.Uid(ctx)
	)

	// 检查commentId
	comment, err := s.CommentBiz.GetComment(ctx, commentId,
		biz.DoNotPopulateExt(), biz.DoNotPopulateImages())
	if err != nil {
		return xerror.Wrapf(err, "comment srv pin comment failed").WithCtx(ctx)
	}

	// 不能对非主评论进行置顶操作
	if !comment.IsRoot() {
		return xerror.Wrap(global.ErrPinFailNotRoot)
	}

	// oid不匹配不能置顶
	if comment.Oid != oid {
		return xerror.Wrap(global.ErrOidNotMatch)
	}

	// 检查用户是否有权置顶评论
	// 只有oid的作者才可以指定评论
	resp, err := dep.GetNoter().IsUserOwnNote(ctx, &notev1.IsUserOwnNoteRequest{
		Uid:    uid,
		NoteId: comment.Oid,
	})
	if err != nil {
		return xerror.Wrapf(err, "comment biz failed to check owner").WithExtra("commentId", comment.Oid).WithCtx(ctx)
	}

	if !resp.GetResult() {
		return xerror.Wrap(global.ErrYouCantPinComment).WithExtras("commentId", comment.Oid, "uid", uid).WithCtx(ctx)
	}

	err = s.CommentInteractBiz.PinComment(ctx, oid, commentId, action)
	if err != nil {
		return xerror.Wrapf(err, "comment srv failed to do pin comment").
			WithCtx(ctx).
			WithExtras("commentId", commentId, "action", action, "oid", oid)
	}
	return nil
}

// 分页获取主评论
func (s *CommentSrv) PageGetRootComments(ctx context.Context, oid, cursor int64, sortBy int8) (*model.PageComments, error) {
	const (
		want = 18
	)

	rootComments, err := s.CommentBiz.GetRootComments(ctx, oid, cursor, want, sortBy)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to get root comments").WithCtx(ctx)
	}

	err = s.CommentInteractBiz.PopulateLikes(ctx, rootComments.Items)
	if err != nil {
		xlog.Msg("comment srv failed to populate root comments").Extras("oid", oid, "cursor", cursor).Errorx(ctx)
	}

	return rootComments, nil
}

// 分页获取子评论
func (s *CommentSrv) PageGetSubComments(ctx context.Context, oid, rootId int64, cursor int64) (*model.PageComments, error) {
	const (
		want = 4
	)

	subComments, err := s.CommentBiz.GetSubComments(ctx, oid, rootId, want, cursor)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to get sub comments").WithCtx(ctx)
	}

	err = s.CommentInteractBiz.PopulateLikes(ctx, subComments.Items)
	if err != nil {
		xlog.Msg("comment srv failed to populate sub comments").Extras("oid", oid, "cursor", cursor).Errorx(ctx)
	}

	return subComments, nil
}

// 按照指定分页页码获取子评论
func (s *CommentSrv) PageListSubComments(ctx context.Context, oid, rootId int64, page, count int) ([]*model.CommentItem, int64, error) {
	lgExts := []any{"oid", oid, "root_id", rootId}
	subComments, total, err := s.CommentBiz.GetSubCommentsByPage(ctx, oid, rootId, page, count)
	if err != nil {
		return nil, 0, xerror.Wrapf(err, "comment srv failed to get subs").
			WithExtras(lgExts).WithCtx(ctx)
	}

	err = s.CommentInteractBiz.PopulateLikes(ctx, subComments)
	if err != nil {
		xlog.Msg("comment srv failed to populate sub comments").Extras(lgExts).Errorx(ctx)
	}

	return subComments, total, nil
}

// 获取对象的评论，包含主评论及其下的子评论
func (s *CommentSrv) PageGetObjectComments(ctx context.Context, oid, cursor int64, sortBy int8) (
	*model.PageDetailedComments, error,
) {

	// 先拿主评论
	roots, err := s.PageGetRootComments(ctx, oid, cursor, sortBy)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to get object replies").
			WithCtx(ctx).WithExtras("oid", oid, "cursor", cursor, "sortBy", sortBy)
	}

	// 获取子评论
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)
	var subs = make([]*model.PageComments, len(roots.Items))
	for i, root := range roots.Items {
		idx, r := i, root // prevent for-loop issue
		eg.Go(func() error {
			return recovery.Do(func() error {
				sub, egErr := s.PageGetSubComments(ctx, oid, r.Id, 0)
				if egErr != nil {
					return xerror.Wrapf(egErr, "goroutine page get sub-replies failed").
						WithExtras("rootId", r.Id, "oid", oid).WithCtx(ctx)
				}

				subs[idx] = sub
				return nil
			})
		})
	}

	err = eg.Wait()
	if err != nil {
		// 服务获取子评论
		return nil, xerror.Wrapf(err, "comment srv failed to get sub replies for root").WithCtx(ctx)
	}

	// 拼装结果
	comments := make([]*model.DetailedCommentItem, 0, len(roots.Items))
	for i, root := range roots.Items {
		comments = append(comments, &model.DetailedCommentItem{
			Root: root,
			Subs: subs[i],
		})
	}
	ret := model.PageDetailedComments{
		Items:      comments,
		NextCursor: roots.NextCursor,
		HasNext:    roots.HasNext,
	}

	return &ret, nil
}

func (s *CommentSrv) PageGetObjectCommentsV2(ctx context.Context, oid, cursor int64, sortBy int8) (
	*model.PageDetailedCommentsV2, error,
) {
	// 先拿主评论
	roots, err := s.PageGetRootComments(ctx, oid, cursor, sortBy)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to get object replies").
			WithCtx(ctx).WithExtras("oid", oid, "cursor", cursor, "sortBy", sortBy)
	}

	// 获取子评论
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)

	var subs = make([]*model.PageCommentsWithTotal, len(roots.Items))
	for i, root := range roots.Items {
		idx, r := i, root // prevent for-loop issue
		eg.Go(func() error {
			return recovery.Do(func() error {
				// 每条主评论默认展示第一页的5条子评论
				sub, total, egErr := s.PageListSubComments(ctx, oid, r.Id, 1, 5)
				if egErr != nil {
					return xerror.Wrapf(egErr, "goroutine page get sub-replies failed").
						WithExtras("rootId", r.Id, "oid", oid).WithCtx(ctx)
				}

				subs[idx] = &model.PageCommentsWithTotal{
					Items: sub,
					Total: total}

				return nil
			})
		})
	}

	err = eg.Wait()
	if err != nil {
		// 服务获取子评论
		return nil, xerror.Wrapf(err, "comment srv failed to get sub replies for root").WithCtx(ctx)
	}

	// 拼装结果
	comments := make([]*model.DetailedCommentItemV2, 0, len(roots.Items))
	for i, root := range roots.Items {
		comments = append(comments, &model.DetailedCommentItemV2{
			Root: root,
			Subs: subs[i],
		})
	}
	ret := model.PageDetailedCommentsV2{
		Items:      comments,
		NextCursor: roots.NextCursor,
		HasNext:    roots.HasNext,
	}

	return &ret, nil
}

// 获取置顶评论
func (s *CommentSrv) GetPinnedComment(ctx context.Context, oid int64) (*model.DetailedCommentItem, error) {
	// 先找出置顶主评论
	root, err := s.CommentBiz.GetPinnedComment(ctx, oid)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv get pinned comment failed")
	}

	if err = s.CommentInteractBiz.PopulateLike(ctx, root); err != nil {
		xlog.Msg("comment srv failed to populate pinned comment").Errorx(ctx)
	}

	// 获取对应子评论
	subs, err := s.CommentBiz.GetSubComments(ctx, oid, root.Id, 10, 0)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to get sub replies for pinned comment").WithCtx(ctx)
	}

	if err := s.CommentInteractBiz.PopulateLikes(ctx, subs.Items); err != nil {
		xlog.Msg("comment srv failed to populate pinned comment's sub replies").Errorx(ctx)
	}

	return &model.DetailedCommentItem{
		Root: root,
		Subs: subs,
	}, nil
}

// 获取评论数量
func (s *CommentSrv) GetCommentCount(ctx context.Context, oid int64) (int64, error) {
	cnt, err := s.CommentBiz.CountComment(ctx, oid)
	if err != nil {
		return 0, xerror.Wrapf(err, "comment srv failed to count comment").WithExtra("oid", oid).WithCtx(ctx)
	}

	return cnt, nil
}

// 获取评论点赞数量
func (s *CommentSrv) GetCommentLikesCount(ctx context.Context, commentId int64) (int64, error) {
	cnt, err := s.CommentInteractBiz.CountCommentLikes(ctx, commentId)
	if err != nil {
		return 0, xerror.Wrapf(err, "comment srv failed to get comment likes count").WithExtra("commentId", commentId).WithCtx(ctx)
	}

	return cnt, nil
}

// 获取评论点踩数量
func (s *CommentSrv) GetCommentDislikesCount(ctx context.Context, commentId int64) (int64, error) {
	cnt, err := s.CommentInteractBiz.CountCommentDislikes(ctx, commentId)
	if err != nil {
		return 0, xerror.Wrapf(err, "comment srv failed to get comment dislikes count").
			WithExtra("commentId", commentId).WithCtx(ctx)
	}

	return cnt, nil
}

// 检查用户是否发起了评论
func (s *CommentSrv) CheckUserIsReplied(ctx context.Context, uid int64, oid int64) (bool, error) {
	ok, err := s.CommentBiz.CheckUserIsCommented(ctx, uid, oid)
	if err != nil {
		return false, xerror.Wrapf(err, "comment srv failed to check user replied on").
			WithExtras("uid", uid, "oid", oid).
			WithCtx(ctx)
	}

	return ok, nil
}

// 获取多个对象的评论数
func (s *CommentSrv) BatchGetCountComment(ctx context.Context, oids []int64) (map[int64]int64, error) {
	if len(oids) == 0 {
		return nil, xerror.ErrArgs.Msg("invalid number of oids")
	}

	cnts, err := s.CommentBiz.BatchCountComment(ctx, oids)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to batch count comment").WithCtx(ctx)
	}

	return cnts, nil
}

func (s *CommentSrv) BatchCheckUserIsReplied(ctx context.Context, uidOids map[int64][]int64) ([]model.UidCommentOnOid, error) {
	commted, err := s.CommentBiz.BatchCheckUserIsCommented(ctx, uidOids)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to batch check user replied on").WithCtx(ctx)
	}

	return commted, nil
}

func (s *CommentSrv) BatchCheckUserLikeStatus(ctx context.Context, uidCommentIds map[int64][]int64) (
	map[int64][]biz.CommentLikeStatus, error,
) {
	resp, err := s.CommentInteractBiz.BatchCheckUserLikeStatus(ctx, uidCommentIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to batch check user like status").WithCtx(ctx)
	}

	return resp, err
}

func (s *CommentSrv) GetCommentImagesUploadAuth(ctx context.Context, cnt int32) (*biz.ImageAuth, error) {
	return s.AssetManagerBiz.BatchGetImageAuths(ctx, cnt)
}

// 检查评论是否存在
func (s *CommentSrv) BatchCheckCommentExist(ctx context.Context, ids []int64) (map[int64]bool, error) {
	ids = xslice.Uniq(ids)

	result := make(map[int64]bool, len(ids))
	if len(ids) == 0 {
		return result, nil
	}

	exists, err := s.CommentBiz.BatchGetComment(ctx, ids,
		biz.DoNotPopulateExt(), biz.DoNotPopulateImages())
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv batch get comment failed").WithCtx(ctx)
	}

	m := xslice.MakeMap(exists, func(v *model.CommentItem) int64 { return v.Id })
	for _, id := range ids {
		if _, ok := m[id]; ok {
			result[id] = true
		} else {
			result[id] = false
		}
	}

	return result, nil
}
