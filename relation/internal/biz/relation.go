package biz

import (
	"context"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/relation/internal/global"
	"github.com/ryanreadbooks/whimer/relation/internal/infra"
	"github.com/ryanreadbooks/whimer/relation/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/relation/internal/model"

	uslices "github.com/ryanreadbooks/whimer/misc/utils/slices"
)

// 关注相关
type RelationBiz interface {
	// follower对followee的关注
	UserFollow(ctx context.Context, follower, followee uint64) error
	// follower取消对followee关注
	UserUnFollow(ctx context.Context, follower, followee uint64) error
	// 获取用户的关注列表
	GetUserFollowingList(ctx context.Context, uid uint64, offset uint64, limit int) ([]uint64, model.ListResult, error)
	// 获取用户的粉丝列表
	GetUserFansList(ctx context.Context, uid uint64, offset uint64, limit int) ([]uint64, model.ListResult, error)
	// 获取用户关注数
	GetUserFollowingCount(ctx context.Context, uid uint64) (uint64, error)
	// 获取用户粉丝数
	GetUserFanCount(ctx context.Context, uid uint64) (uint64, error)
	// 检查用户是否关注了某些人
	BatchCheckUserFollowStatus(ctx context.Context, uid uint64, others []uint64) (map[uint64]bool, error)
	// 检查用户是否关注了某个人
	CheckUserFollowStatus(ctx context.Context, uid, other uint64) (bool, error)
}

type relationBiz struct {
}

func NewRelationBiz() RelationBiz {
	b := &relationBiz{}

	return b
}

// follower发起对followee的关注
func (b *relationBiz) UserFollow(ctx context.Context, follower, followee uint64) error {
	err := infra.Dao().DB().Transact(ctx, func(ctx context.Context) error {
		// 需要先检查当前两人的关注状态
		cur, err := infra.Dao().RelationDao.FindByAlphaBeta(ctx, follower, followee, true)
		if err != nil {
			if !errors.Is(err, xsql.ErrNoRecord) {
				return xerror.Wrapf(err, "dao find by alpha and beta failed")
			}
		}

		var (
			now      = time.Now().Unix()
			relation = &dao.Relation{
				UserAlpha: follower,
				UserBeta:  followee,
				Actime:    now,
				Amtime:    now,
			}
		)

		if cur == nil {
			// 两者没有关注关系
			relation.Link = dao.LinkForward
		} else {
			// 两者有过关注关系
			if (cur.UserAlpha == follower && cur.Link == dao.LinkForward) ||
				(cur.UserBeta == follower && cur.Link == dao.LinkBackward) ||
				(cur.Link == dao.LinkMutual) {
				// 无需重复关注
				return global.ErrAlreadyFollow
			} else {
				if (cur.UserAlpha == followee && cur.Link == dao.LinkForward) ||
					(cur.UserBeta == followee && cur.Link == dao.LinkBackward) {
					// followee 已经对 follower发起了关注，此时就需要改成相互关注状态
					relation.Link = dao.LinkMutual
				} else {
					relation.Link = dao.LinkForward
				}
			}
			// 注意时间
			if cur.UserAlpha == follower {
				relation.Actime = cur.Actime
				relation.Bctime = cur.Bctime
				relation.Bmtime = cur.Bmtime
			} else {
				relation.Actime = cur.Bctime
				relation.Bctime = cur.Actime
				relation.Bmtime = cur.Amtime
			}
			if relation.Actime == 0 {
				relation.Actime = now
			}
		}

		err = infra.Dao().RelationDao.Insert(ctx, relation)
		if err != nil {
			return xerror.Wrapf(err, "dao insert failed when user follow")
		}

		return nil
	})

	if err != nil {
		return xerror.Wrapf(err, "relation biz follow user transact failed").WithExtras("follower", follower, "followee", followee)
	}

	return nil
}

func (b *relationBiz) UserUnFollow(ctx context.Context, follower, followee uint64) error {
	err := infra.Dao().DB().Transact(ctx, func(ctx context.Context) error {
		cur, err := infra.Dao().RelationDao.FindByAlphaBeta(ctx, follower, followee, true)
		if err != nil {
			if !errors.Is(err, xsql.ErrNoRecord) {
				return xerror.Wrapf(err, "dao find by alpha and beta failed")
			}
		}

		var (
			now      = time.Now().Unix()
			relation = &dao.Relation{
				UserAlpha: follower,
				UserBeta:  followee,
				Actime:    now,
				Amtime:    now,
			}
		)

		// 原本没有关注关系 无需处理
		if cur == nil || cur.Link == dao.LinkVacant {
			return nil
		}

		if (cur.UserAlpha == follower && cur.Link == dao.LinkForward) ||
			(cur.UserBeta == follower && cur.Link == dao.LinkBackward) {
			// 原本只是follower的单向关注，此时只需要将链接断开
			relation.Link = dao.LinkVacant
		} else if cur.Link == dao.LinkMutual {
			// 原本是follower和followee互相关注，此时follower取消关注，结果变成followee单向关注
			relation.Link = dao.LinkBackward
		} else {
			// 原本是followee对follower的单向关注
			relation.Link = cur.Link
		}
		// 注意时间
		if cur.UserAlpha == follower {
			relation.Actime = cur.Actime
			relation.Bctime = cur.Bctime
			relation.Bmtime = cur.Bmtime
		} else {
			relation.Actime = cur.Bctime
			relation.Bctime = cur.Actime
			relation.Bmtime = cur.Amtime
		}
		if relation.Actime == 0 {
			relation.Actime = now
		}

		err = infra.Dao().RelationDao.Insert(ctx, relation)
		if err != nil {
			return xerror.Wrapf(err, "dao insert failed when user unfollow")
		}

		return nil
	})

	if err != nil {
		return xerror.Wrapf(err, "relation biz unfollow user transact failed").WithExtras("follower", follower, "followee", followee)
	}

	return nil
}

// 获取用户的关注列表
func (b *relationBiz) GetUserFollowingList(ctx context.Context, uid uint64, offset uint64, limit int) ([]uint64, model.ListResult, error) {
	var lr model.ListResult
	followings, next, more, err := infra.Dao().RelationDao.FindUidLinkTo(ctx, uid, offset, limit)
	if err != nil {
		return nil, lr, xerror.Wrapf(err, "relation biz failed to find uid link to").WithExtras("offset", offset, "limit", limit)
	}

	lr.NextOffset = next
	lr.HasMore = more
	return followings, lr, nil
}

// 获取用户的粉丝列表
func (b *relationBiz) GetUserFansList(ctx context.Context, uid uint64, offset uint64, limit int) ([]uint64, model.ListResult, error) {
	var lr model.ListResult
	fans, next, more, err := infra.Dao().RelationDao.FindUidGotLinked(ctx, uid, offset, limit)
	if err != nil {
		return nil, lr, xerror.Wrapf(err, "relation biz failed to find uid got linked to").WithExtras("offset", offset, "limit", limit)
	}

	lr.NextOffset = next
	lr.HasMore = more
	return fans, lr, nil
}

// 获取用户关注数
func (b *relationBiz) GetUserFollowingCount(ctx context.Context, uid uint64) (uint64, error) {
	cnt, err := infra.Dao().RelationDao.CountUidFollowings(ctx, uid)
	if err != nil {
		return 0, xerror.Wrapf(err, "relation biz failed to count user followings").WithCtx(ctx)
	}

	return cnt, nil
}

// 获取用户粉丝数
func (b *relationBiz) GetUserFanCount(ctx context.Context, uid uint64) (uint64, error) {
	cnt, err := infra.Dao().RelationDao.CountUidFans(ctx, uid)
	if err != nil {
		return 0, xerror.Wrapf(err, "relation biz failed to count user fans").WithCtx(ctx)
	}

	return cnt, nil
}

// 检查uid是否关注了others
func (b *relationBiz) BatchCheckUserFollowStatus(ctx context.Context, uid uint64, others []uint64) (map[uint64]bool, error) {
	followings, err := infra.Dao().RelationDao.FindAllUidLinkTo(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "relation biz failed to find all followings").WithCtx(ctx)
	}

	res := make(map[uint64]bool, len(others))
	fm := uslices.AsMap(followings)
	for _, o := range others {
		if _, ok := fm[o]; ok {
			res[o] = true
		}
	}

	return res, nil
}

// 检查uid是否关注了other
func (b *relationBiz) CheckUserFollowStatus(ctx context.Context, uid, other uint64) (bool, error) {
	rel, err := infra.Dao().RelationDao.FindByAlphaBeta(ctx, uid, other, false)
	if err != nil {
		return false, xerror.Wrapf(err, "relation biz failed to find by alpha-beta").
			WithExtras("uid", uid, "other", other).
			WithCtx(ctx)
	}

	relLink := rel.Link
	if relLink == dao.LinkMutual {
		return true, nil
	}

	if rel.UserAlpha == uid {
		return relLink == dao.LinkForward, nil
	} else {
		// beta == uid
		return relLink == dao.LinkBackward, nil
	}
}
