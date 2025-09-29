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
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xnet"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
)

type CommentBiz struct{}

// 评论基础功能领域
func NewCommentBiz() CommentBiz {
	return CommentBiz{}
}

func (b *CommentBiz) addCommentAssets(ctx context.Context, newCommentId int64, req *model.AddCommentReq) error {
	if req.Type == model.CommentImageText {
		// 插入评论图片资源
		assets := makeCommentAssetPO(newCommentId, req.Images)
		err := infra.Dao().CommentAssetDao.BatchInsert(ctx, assets)
		if err != nil {
			return xerror.Wrapf(err, "comment biz batch insert image assets failed")
		}
	}

	return nil
}

// 用户发表评论
func (b *CommentBiz) AddComment(ctx context.Context, req *model.AddCommentReq) (*model.AddCommentRes, error) {
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

	var newCommentId int64

	now := time.Now().Unix()
	newComment := dao.Comment{
		Oid:      oid,
		Type:     int8(req.Type),
		Content:  req.Content,
		Uid:      uid,
		RootId:   rootId,
		ParentId: parentId,
		ReplyUid: req.ReplyUid,
		State:    int8(model.CommentStateNormal),
		Ip:       ip,
		Ctime:    now,
		Mtime:    now,
	}

	err = infra.Dao().Transact(ctx, func(ctx context.Context) error {
		if model.IsRoot(rootId, parentId) {
			// 新增的是主评论 直接新增
			newCommentId, err = infra.Dao().CommentDao.Insert(ctx, &newComment)
			if err != nil {
				return xerror.Wrapf(err, "comment biz insert root comment failed")
			}

			return b.addCommentAssets(ctx, newCommentId, req)
		} else {
			// 新增的是评论的评论 插入前校验能否插入
			// 检查被评论的评论是否存在
			err := b.isCommentAddable(ctx, rootId, parentId)
			if err != nil {
				return xerror.Wrapf(err, "isCommentAddable check failed")
			}

			// 可以插入
			newCommentId, err = infra.Dao().CommentDao.Insert(ctx, &newComment)
			if err != nil {
				return xerror.Wrapf(err, "comment biz dao insert comment failed")
			}

			return b.addCommentAssets(ctx, newCommentId, req)
		}
	})

	if err != nil {
		return nil, xerror.Wrapf(err, "comment biz failed to insert comment")
	}

	concurrent.DoneIn(10*time.Second, func(ctx context.Context) {
		infra.Dao().CommentDao.IncrCommentCount(ctx, oid)
	})

	return &model.AddCommentRes{Uid: uid, CommentId: newCommentId}, nil
}

func (b *CommentBiz) findByIdForUpdate(ctx context.Context, commentId int64) (*model.CommentItem, error) {
	c, err := infra.Dao().CommentDao.FindByIdForUpdate(ctx, commentId)
	if err != nil {
		if !xsql.IsNoRecord(err) {
			return nil, xerror.Wrapf(err, "comment biz find by for update failed")
		}
		return nil, xerror.Wrap(global.ErrCommentNotFound)
	}

	return model.NewCommentItemFromDao(c), nil
}

// 检查是否能够发布子评论
func (b *CommentBiz) isCommentAddable(ctx context.Context, rootId, parentId int64) error {
	// 两个都需要检查
	if rootId != 0 {
		root, err := b.findByIdForUpdate(ctx, rootId)
		if err != nil {
			return xerror.Wrap(err)
		}
		// 确保root真的是root
		if !root.IsRoot() {
			return xerror.Wrap(global.ErrRootCommentIsNotRoot)
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
func (b *CommentBiz) DelComment(ctx context.Context, oid, commentId int64) error {
	var (
		uid = metadata.Uid(ctx)
	)

	// 检查评论是否存在
	existingComment, err := b.GetComment(ctx, commentId)
	if err != nil {
		return xerror.Wrapf(err, "comment biz failed to get comment")
	}
	if existingComment.Oid != oid {
		return xerror.Wrap(global.ErrOidNotMatch)
	}

	if err := b.isCommentDeletable(ctx, uid, existingComment); err != nil {
		return xerror.Wrapf(err, "comment biz check comment is not deletable")
	}

	// 开始删除
	// 删除的如果是主评论 需要一并删除所有子评论
	// 否则就只删除评论本身
	// 删除子评论的子评论也只是只删除评论本身
	err = infra.Dao().Transact(ctx, func(ctx context.Context) error {
		if model.IsRoot(existingComment.RootId, existingComment.ParentId) {
			_, err := b.findByIdForUpdate(ctx, commentId)
			if err != nil {
				return xerror.Wrapf(err, "comment biz dao find comment for update failed")
			}

			// 删除主评论
			err = infra.Dao().CommentDao.DeleteById(ctx, commentId)
			if err != nil {
				return xerror.Wrapf(err, "comment biz dao delete root by id failed")
			}

			// 删除评论下的资源
			err = infra.Dao().CommentAssetDao.BatchDeleteByRoot(ctx, commentId)
			if err != nil {
				return xerror.Wrapf(err, "comment biz dao delete by root failed")
			}

			// 删除其下子评论
			err = infra.Dao().CommentDao.DeleteByRoot(ctx, commentId)
			if err != nil {
				return xerror.Wrapf(err, "comment biz dao delete comments by rootid failed")
			}

			return nil
		} else {
			if err = infra.Dao().CommentDao.DeleteById(ctx, commentId); err != nil {
				return xerror.Wrapf(err, "comment biz failed to delete")
			}
		}
		return nil
	})

	if err != nil {
		return xerror.Wrapf(err, "comment biz failed to delete comment")
	}

	// 缓存中减少一个
	concurrent.DoneIn(10*time.Second, func(ctx context.Context) {
		infra.Dao().CommentDao.DecrCommentCount(ctx, existingComment.Oid)
	})
	return nil
}

// 检查评论是否存在
func (b *CommentBiz) GetComment(ctx context.Context, commentId int64) (*model.CommentItem, error) {
	comment, err := infra.Dao().CommentDao.FindById(ctx, commentId)
	if err != nil {
		if !xsql.IsNoRecord(err) {
			return nil, xerror.Wrapf(err, "comment biz failed to get").WithExtra("cid", commentId).WithCtx(ctx)
		}

		return nil, xerror.Wrap(global.ErrCommentNotFound)
	}

	item := model.NewCommentItemFromDao(comment)

	if err = b.PopulateCommentImages(ctx, []*model.CommentItem{item}); err != nil {
		return nil, xerror.Wrapf(err, "comment biz failed to populate images")
	}

	return item, nil
}

// 检查是否可以删除评论, 比如用户权限判断
func (b *CommentBiz) isCommentDeletable(ctx context.Context, uid int64, item *model.CommentItem) error {
	var (
		owner = item.Uid
	)

	// 用户是评论的作者可以删除
	if uid == owner {
		return nil
	}

	// 用户是评论对象的作者可以删除
	resp, err := dep.GetNoter().IsUserOwnNote(ctx, &notev1.IsUserOwnNoteRequest{
		Uid:    uid,
		NoteId: item.Oid,
	})
	if err != nil {
		return xerror.Wrapf(err, "comment biz failed to check owner").WithExtra("oid", item.Oid).WithCtx(ctx)
	}

	if !resp.GetResult() {
		return xerror.Wrap(global.ErrYouDontOwnThis)
	}

	return nil
}

// 仅获取根评论, 不获取根评论下的子评论
// 每次返回10条
func (b *CommentBiz) GetRootComments(ctx context.Context, oid int64, cursor int64, want int, sortBy int8) (*model.PageComments, error) {
	if want <= 0 {
		want = 18
	}

	data, err := infra.Dao().CommentDao.GetRoots(ctx, oid, cursor, want)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment biz failed to get roots").WithCtx(ctx).
			WithExtras("oid", oid, "cursor", cursor)
	}

	// 计算下一个cursor
	dataLen := len(data)
	var nextCursor int64 = 0
	hasNext := dataLen == want
	if dataLen > 0 {
		nextCursor = data[dataLen-1].Id
	}
	if !hasNext {
		nextCursor = 0
	}

	items := make([]*model.CommentItem, 0, dataLen)
	rootIds := make([]int64, 0, dataLen)
	for _, item := range data {
		items = append(items, model.NewCommentItemFromDao(item))
		rootIds = append(rootIds, item.Id)
	}

	// 填充主评论的子评论数量
	err = b.PopulateSubCommentsCount(ctx, items)
	if err != nil {
		// 获取子评论失败不返回错误
		xlog.Msg("comment biz batch count sub comments failed").Extras("rootIds", rootIds).Errorx(ctx)
	}

	err = b.PopulateCommentImages(ctx, items)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment biz failed to populate images").
			WithExtra("rootIds", rootIds).WithCtx(ctx)
	}

	return &model.PageComments{
		Items:      items,
		NextCursor: nextCursor,
		HasNext:    hasNext,
	}, nil
}

// 仅获取子评论
// 获取对象oid中rootId评论下的子评论
// 每次返回5条
func (b *CommentBiz) GetSubComments(ctx context.Context, oid, rootId int64, want int, cursor int64) (*model.PageComments, error) {
	if want <= 0 {
		want = 10
	}

	data, err := infra.Dao().CommentDao.GetSubReplies(ctx, oid, rootId, cursor, want)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment biz failed to get sub comments").
			WithExtras("oid", oid, "rootId", rootId, "cursor", cursor).WithCtx(ctx)
	}
	dataLen := len(data)
	var nextCursor int64 = 0
	hasNext := dataLen == want
	if dataLen > 0 {
		nextCursor = data[dataLen-1].Id
	}
	if !hasNext {
		nextCursor = 0
	}

	items := make([]*model.CommentItem, 0, dataLen)
	for _, item := range data {
		items = append(items, model.NewCommentItemFromDao(item))
	}

	if err := b.PopulateCommentImages(ctx, items); err != nil {
		return nil, xerror.Wrapf(err, "comment biz failed to populate images").
			WithExtras("rootId", rootId, "oid", oid).WithCtx(ctx)
	}

	return &model.PageComments{
		Items:      items,
		NextCursor: nextCursor,
		HasNext:    hasNext,
	}, nil
}

// 按照页码获取子评论
func (b *CommentBiz) GetSubCommentsByPage(ctx context.Context, oid, rootId int64, page, cnt int) (
	[]*model.CommentItem, int64, error) {

	total, err := infra.Dao().CommentDao.CountSubs(ctx, oid, rootId)
	if err != nil {
		return nil, 0, xerror.Wrapf(err, "comment dao failed to count subs").WithCtx(ctx)
	}

	data, err := infra.Dao().CommentDao.PageGetSubs(ctx, oid, rootId, page, cnt)
	if err != nil {
		return nil, 0, xerror.Wrapf(err, "comment dao failed to page get subs").WithCtx(ctx)
	}
	items := make([]*model.CommentItem, 0, len(data))
	for _, item := range data {
		items = append(items, model.NewCommentItemFromDao(item))
	}

	if err := b.PopulateCommentImages(ctx, items); err != nil {
		return nil, 0, xerror.Wrapf(err, "comment biz failed to populate images").
			WithExtras("rootId", rootId, "oid", oid).WithCtx(ctx)
	}

	return items, total, nil
}

func (b *CommentBiz) CountComment(ctx context.Context, oid int64) (int64, error) {
	cnt, err := infra.Dao().CommentDao.CountByOid(ctx, oid)
	if err != nil {
		return 0, xerror.Wrapf(err, "comment biz failed to get comments count").WithCtx(ctx).WithExtra("oid", oid)
	}

	return cnt, nil
}

// 批量获取评论数量
func (b *CommentBiz) BatchCountComment(ctx context.Context, oids []int64) (map[int64]int64, error) {
	if len(oids) == 0 {
		return nil, xerror.ErrArgs.Msg("invalid number of oids")
	}
	resp, err := infra.Dao().CommentDao.BatchCountByOid(ctx, oids)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment biz failed to batch get comments count").WithCtx(ctx)
	}

	return resp, nil
}

// 检查用户是否发起了评论
func (b *CommentBiz) CheckUserIsCommented(ctx context.Context, uid int64, oid int64) (bool, error) {
	cnt, err := infra.Dao().CommentDao.CountByOidUid(ctx, oid, uid)
	if err != nil {
		return false, xerror.Wrapf(err, "comment biz failed to check user is commented on object").
			WithExtras("uid", uid, "oid", oid).
			WithCtx(ctx)
	}

	return cnt != 0, nil
}

// 批量检查用户是否发起了评论
// uid => [oid1, oid2, ..., oidN]
func (b *CommentBiz) BatchCheckUserIsCommented(ctx context.Context, uidOids map[int64][]int64) ([]model.UidCommentOnOid, error) {
	resp, err := infra.Dao().CommentDao.FindByUidsOids(ctx, uidOids)
	if err != nil {
		return nil, xerror.Wrapf(err, "comment biz failed to batch check user is commented").WithCtx(ctx)
	}

	// 记录uid评论过哪些oids
	commented := make(map[int64][]int64, len(resp))
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

// 获取置顶评论
func (b *CommentBiz) GetPinnedComment(ctx context.Context, oid int64) (*model.CommentItem, error) {
	pinned, err := infra.Dao().CommentDao.GetPinned(ctx, oid)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return nil, global.ErrNoPinComment
		}

		return nil, xerror.Wrapf(err, "comment biz get pinned failed").WithExtra("oid", oid).WithCtx(ctx)
	}

	item := model.NewCommentItemFromDao(pinned)
	if err := b.PopulateCommentImages(ctx, []*model.CommentItem{item}); err != nil {
		return nil, xerror.Wrapf(err, "comment biz failed to populate images").
			WithExtras("rootId", item.RootId, "oid", oid).WithCtx(ctx)
	}

	return item, nil
}

// 填充评论的子评论数量(只对主评论生效)
func (b *CommentBiz) PopulateSubCommentsCount(ctx context.Context, items []*model.CommentItem) error {
	rootIds := make([]int64, 0, len(items))
	for _, r := range items {
		rootIds = append(rootIds, r.Id)
	}
	if len(rootIds) == 0 {
		return nil
	}

	resp, err := infra.Dao().CommentDao.BatchCountSubs(ctx, rootIds)
	if err != nil {
		return xerror.Wrapf(err, "comment biz failed to batch count subs").
			WithExtra("roots", rootIds).
			WithCtx(ctx)
	}

	for _, item := range items {
		item.SubsCount = resp[item.Id]
	}

	return nil
}

// 填充评论的图片资源
func (b *CommentBiz) PopulateCommentImages(ctx context.Context, items []*model.CommentItem) error {
	if len(items) == 0 {
		return nil
	}

	commentIds := make([]int64, 0, len(items))
	for _, r := range items {
		if r.Type != int8(model.CommentText) {
			commentIds = append(commentIds, r.Id)
		}
	}

	commentIds = xslice.Uniq(commentIds)
	if len(commentIds) == 0 {
		return nil
	}

	assetsMap, err := infra.Dao().CommentAssetDao.BatchGetByCommentIds(ctx, commentIds)
	if err != nil {
		return xerror.Wrapf(err, "comment biz failed to batch get comment assets").
			WithExtra("comment_ids", commentIds).
			WithCtx(ctx)
	}

	// 填充图片资源
	for _, item := range items {
		if item.Type != int8(model.CommentText) {
			if assets, ok := assetsMap[item.Id]; ok {
				item.Images = makePbCommentImage(assets)
			}
		}
	}

	return nil
}
