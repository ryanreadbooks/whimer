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
)

// 关注相关
type RelationBiz interface {
	// follower对followee的关注
	UserFollow(ctx context.Context, follower, followee uint64) error
	// follower取消对followee关注
	UserUnFollow(ctx context.Context, follower, followee uint64) error
	// 获取用户的关注列表
	GetUserFollowingList(ctx context.Context, uid uint64, offset uint64, limit int) ([]uint64, error)
	// 获取用户的粉丝列表
	GetUserFansList(ctx context.Context, uid uint64, offset uint64, limit int) ([]uint64, error)
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

func (b *relationBiz) GetUserFollowingList(ctx context.Context, uid uint64, offset uint64, limit int) ([]uint64, error) {

	return nil, nil
}

func (b *relationBiz) GetUserFansList(ctx context.Context, uid uint64, offset uint64, limit int) ([]uint64, error) {

	return nil, nil
}
