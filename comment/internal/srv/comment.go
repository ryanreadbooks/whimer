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
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"

	"golang.org/x/sync/errgroup"
)

type CommentSrv struct {
	CommentBiz         biz.CommentBiz
	CommentInteractBiz biz.CommentInteractBiz
}

func NewCommentSrv(s *Service, biz biz.Biz) *CommentSrv {
	return &CommentSrv{
		CommentBiz:         biz.CommentBiz,
		CommentInteractBiz: biz.CommentInteractBiz,
	}
}

// 用户发表评论
func (s *CommentSrv) AddReply(ctx context.Context, req *model.AddReplyReq) (*model.AddReplyRes, error) {
	res, err := s.CommentBiz.AddReply(ctx, req)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to add reply").WithCtx(ctx).WithExtra("req", req)
	}

	// TODO 通知被评论的用户

	return res, nil
}

// 用户删除评论
func (s *CommentSrv) DelReply(ctx context.Context, rid uint64) error {
	err := s.CommentBiz.DelReply(ctx, rid)
	if err != nil {
		return xerror.Wrapf(err, "comment srv failed to del reply").
			WithCtx(ctx).
			WithExtra("rid", rid)
	}

	return nil
}

// 用户点赞/取消点赞某条评论
func (s *CommentSrv) LikeReply(ctx context.Context, rid uint64, action int8) error {
	err := s.CommentInteractBiz.LikeReply(ctx, rid, action)
	if err != nil {
		return xerror.Wrapf(err, "comment srv failed to do like reply").
			WithCtx(ctx).
			WithExtras("rid", rid, "action", action)
	}

	return nil
}

// 用户点踩/取消点踩某条评论
func (s *CommentSrv) DislikeReply(ctx context.Context, rid uint64, action int8) error {
	err := s.CommentInteractBiz.DislikeReply(ctx, rid, action)
	if err != nil {
		return xerror.Wrapf(err, "comment srv failed to do dislike reply").
			WithCtx(ctx).
			WithExtras("rid", rid, "action", action)
	}

	return nil
}

// 置顶评论/取消置顶评论
func (s *CommentSrv) PinReply(ctx context.Context, oid, rid uint64, action int8) error {
	var (
		uid = metadata.Uid(ctx)
	)

	// 检查rid
	reply, err := s.CommentBiz.GetReply(ctx, rid)
	if err != nil {
		return xerror.Wrapf(err, "comment srv pin reply failed").WithCtx(ctx)
	}

	// 不能对非主评论进行置顶操作
	if !reply.IsRoot() {
		return xerror.Wrap(global.ErrPinFailNotRoot)
	}

	// oid不匹配不能置顶
	if reply.Oid != oid {
		return xerror.Wrap(global.ErrOidNotMatch)
	}

	// 检查用户是否有权置顶评论
	// 只有oid的作者才可以指定评论
	resp, err := dep.GetNoter().IsUserOwnNote(ctx, &notev1.IsUserOwnNoteRequest{
		Uid:    uid,
		NoteId: reply.Oid,
	})
	if err != nil {
		return xerror.Wrapf(err, "comment biz failed to check owner").WithExtra("replyId", reply.Oid).WithCtx(ctx)
	}

	if !resp.GetResult() {
		return xerror.Wrap(global.ErrYouCantPinReply).WithExtras("replyId", reply.Oid, "uid", uid).WithCtx(ctx)
	}

	err = s.CommentInteractBiz.PinReply(ctx, oid, rid, action)
	if err != nil {
		return xerror.Wrapf(err, "comment srv failed to do pin reply").
			WithCtx(ctx).
			WithExtras("rid", rid, "action", action, "oid", oid)
	}
	return nil
}

// 分页获取主评论
func (s *CommentSrv) PageGetRootReplies(ctx context.Context, oid, cursor uint64, sortBy int8) (*model.PageReplies, error) {
	const (
		want = 18
	)

	rootReplies, err := s.CommentBiz.GetRootReplies(ctx, oid, cursor, want, sortBy)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to get root replies").WithCtx(ctx)
	}

	err = s.CommentInteractBiz.PopulateLikes(ctx, rootReplies.Items)
	if err != nil {
		xlog.Msg("comment srv failed to populate root replies").Extras("oid", oid, "cursor", cursor).Errorx(ctx)
	}

	return rootReplies, nil
}

// 分页获取子评论
func (s *CommentSrv) PageGetSubReplies(ctx context.Context, oid, rootId uint64, cursor uint64) (*model.PageReplies, error) {
	const (
		want = 4
	)

	subReplies, err := s.CommentBiz.GetSubReplies(ctx, oid, rootId, want, cursor)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to get sub replies").WithCtx(ctx)
	}

	err = s.CommentInteractBiz.PopulateLikes(ctx, subReplies.Items)
	if err != nil {
		xlog.Msg("comment srv failed to populate sub replies").Extras("oid", oid, "cursor", cursor).Errorx(ctx)
	}

	return subReplies, nil
}

// 按照指定分页页码获取子评论
func (s *CommentSrv) PageListSubReplies(ctx context.Context, oid, rootId uint64, page, count int) ([]*model.ReplyItem, int64, error) {
	lgExts := []any{"oid", oid, "root_id", rootId}
	subReplies, total, err := s.CommentBiz.GetSubRepliesByPage(ctx, oid, rootId, page, count)
	if err != nil {
		return nil, 0, xerror.Wrapf(err, "comment srv failed to get subreplies").
			WithExtras(lgExts).WithCtx(ctx)
	}

	err = s.CommentInteractBiz.PopulateLikes(ctx, subReplies)
	if err != nil {
		xlog.Msg("comment srv failed to populate sub replies").Extras(lgExts).Errorx(ctx)
	}

	return subReplies, total, nil
}

// 获取对象的评论，包含主评论及其下的子评论
func (s *CommentSrv) PageGetObjectReplies(ctx context.Context, oid, cursor uint64, sortBy int8) (
	*model.PageDetailedReplies, error,
) {

	// 先拿主评论
	roots, err := s.PageGetRootReplies(ctx, oid, cursor, sortBy)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to get object replies").
			WithCtx(ctx).WithExtras("oid", oid, "cursor", cursor, "sortBy", sortBy)
	}

	// 获取子评论
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()

	eg, ctx := errgroup.WithContext(ctx)
	var subs = make([]*model.PageReplies, len(roots.Items))
	for i, root := range roots.Items {
		idx, r := i, root // prevent for-loop issue
		eg.Go(func() error {
			return recovery.Do(func() error {
				sub, egErr := s.PageGetSubReplies(ctx, oid, r.Id, 0)
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
	replies := make([]*model.DetailedReplyItem, 0, len(roots.Items))
	for i, root := range roots.Items {
		replies = append(replies, &model.DetailedReplyItem{
			Root: root,
			Subs: subs[i],
		})
	}
	ret := model.PageDetailedReplies{
		Items:      replies,
		NextCursor: roots.NextCursor,
		HasNext:    roots.HasNext,
	}

	return &ret, nil
}

func (s *CommentSrv) PageGetObjectRepliesV2(ctx context.Context, oid, cursor uint64, sortBy int8) (
	*model.PageDetailedRepliesV2, error,
) {
	// 先拿主评论
	roots, err := s.PageGetRootReplies(ctx, oid, cursor, sortBy)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to get object replies").
			WithCtx(ctx).WithExtras("oid", oid, "cursor", cursor, "sortBy", sortBy)
	}

	// 获取子评论
	ctx, cancel := context.WithTimeout(ctx, time.Second*20)
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)

	var subs = make([]*model.PageRepliesWithTotal, len(roots.Items))
	for i, root := range roots.Items {
		idx, r := i, root // prevent for-loop issue
		eg.Go(func() error {
			return recovery.Do(func() error {
				// 每条主评论默认展示第一页的5条子评论
				sub, total, egErr := s.PageListSubReplies(ctx, oid, r.Id, 1, 5)
				if egErr != nil {
					return xerror.Wrapf(egErr, "goroutine page get sub-replies failed").
						WithExtras("rootId", r.Id, "oid", oid).WithCtx(ctx)
				}

				subs[idx] = &model.PageRepliesWithTotal{
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
	replies := make([]*model.DetailedReplyItemV2, 0, len(roots.Items))
	for i, root := range roots.Items {
		replies = append(replies, &model.DetailedReplyItemV2{
			Root: root,
			Subs: subs[i],
		})
	}
	ret := model.PageDetailedRepliesV2{
		Items:      replies,
		NextCursor: roots.NextCursor,
		HasNext:    roots.HasNext,
	}

	return &ret, nil
}

// 获取置顶评论
func (s *CommentSrv) GetPinnedReply(ctx context.Context, oid uint64) (*model.DetailedReplyItem, error) {
	// 先找出置顶主评论
	root, err := s.CommentBiz.GetPinnedReply(ctx, oid)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv get pinned reply failed")
	}

	if err = s.CommentInteractBiz.PopulateLike(ctx, root); err != nil {
		xlog.Msg("comment srv failed to populate pinned reply").Errorx(ctx)
	}

	// 获取对应子评论
	subs, err := s.CommentBiz.GetSubReplies(ctx, oid, root.Id, 10, 0)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to get sub replies for pinned reply").WithCtx(ctx)
	}

	if err := s.CommentInteractBiz.PopulateLikes(ctx, subs.Items); err != nil {
		xlog.Msg("comment srv failed to populate pinned reply's sub replies").Errorx(ctx)
	}

	return &model.DetailedReplyItem{
		Root: root,
		Subs: subs,
	}, nil
}

// 获取评论数量
func (s *CommentSrv) GetReplyCount(ctx context.Context, oid uint64) (uint64, error) {
	cnt, err := s.CommentBiz.CountReply(ctx, oid)
	if err != nil {
		return 0, xerror.Wrapf(err, "comment srv failed to count reply").WithExtra("oid", oid).WithCtx(ctx)
	}

	return cnt, nil
}

// 获取评论点赞数量
func (s *CommentSrv) GetReplyLikesCount(ctx context.Context, rid uint64) (uint64, error) {
	cnt, err := s.CommentInteractBiz.CountReplyLikes(ctx, rid)
	if err != nil {
		return 0, xerror.Wrapf(err, "comment srv failed to get reply likes count").WithExtra("rid", rid).WithCtx(ctx)
	}

	return cnt, nil
}

// 获取评论点踩数量
func (s *CommentSrv) GetReplyDislikesCount(ctx context.Context, rid uint64) (uint64, error) {
	cnt, err := s.CommentInteractBiz.CountReplyDislikes(ctx, rid)
	if err != nil {
		return 0, xerror.Wrapf(err, "comment srv failed to get reply dislikes count").WithExtra("rid", rid).WithCtx(ctx)
	}

	return cnt, nil
}

// 检查用户是否发起了评论
func (s *CommentSrv) CheckUserIsReplied(ctx context.Context, uid int64, oid uint64) (bool, error) {
	ok, err := s.CommentBiz.CheckUserIsReplied(ctx, uid, oid)
	if err != nil {
		return false, xerror.Wrapf(err, "comment srv failed to check user replied on").
			WithExtras("uid", uid, "oid", oid).
			WithCtx(ctx)
	}

	return ok, nil
}

// 获取多个对象的评论数
func (s *CommentSrv) BatchGetCountReply(ctx context.Context, oids []uint64) (map[uint64]uint64, error) {
	if len(oids) == 0 {
		return nil, xerror.ErrArgs.Msg("invalid number of oids")
	}

	cnts, err := s.CommentBiz.BatchCountReply(ctx, oids)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to batch count reply").WithCtx(ctx)
	}

	return cnts, nil
}

func (s *CommentSrv) BatchCheckUserIsReplied(ctx context.Context, uidOids map[int64][]uint64) ([]model.UidCommentOnOid, error) {
	commted, err := s.CommentBiz.BatchCheckUserIsReplied(ctx, uidOids)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to batch check user replied on").WithCtx(ctx)
	}

	return commted, nil
}

func (s *CommentSrv) BatchCheckUserLikeStatus(ctx context.Context, uidReplyIds map[int64][]uint64) (
	map[int64][]biz.ReplyLikeStatus, error,
) {
	resp, err := s.CommentInteractBiz.BatchCheckUserLikeStatus(ctx, uidReplyIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment srv failed to batch check user like status").WithCtx(ctx)
	}

	return resp, err
}
