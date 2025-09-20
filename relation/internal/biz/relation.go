package biz

import (
	"context"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/relation/internal/global"
	"github.com/ryanreadbooks/whimer/relation/internal/infra"
	"github.com/ryanreadbooks/whimer/relation/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/relation/internal/model"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type RelationBiz struct {
}

func NewRelationBiz() *RelationBiz {
	b := &RelationBiz{}

	return b
}

// follower发起对followee的关注
func (b *RelationBiz) UserFollow(ctx context.Context, follower, followee int64) error {
	var (
		now      = time.Now().Unix()
		relation = &dao.Relation{
			UserAlpha: follower,
			UserBeta:  followee,
			Actime:    now,
			Amtime:    now,
		}
	)

	if cachedData, err := infra.Dao().RelationCache.GetLink(ctx, follower, followee); err == nil {
		// 尽量拦截已经关注的状态
		if (cachedData.UserAlpha == follower && cachedData.Link.IsForward()) ||
			(cachedData.UserBeta == follower && cachedData.Link.IsBackward()) ||
			(cachedData.Link.IsMutual()) {
			// 无需重复关注
			return global.ErrAlreadyFollow
		}
	}

	err := infra.Dao().DB().Transact(ctx, func(ctx context.Context) error {
		// 需要先检查当前两人的关注状态
		cur, err := infra.Dao().RelationDao.FindByAlphaBeta(ctx, follower, followee, true)
		if err != nil {
			if !errors.Is(err, xsql.ErrNoRecord) {
				return xerror.Wrapf(err, "dao find by alpha and beta failed")
			}
		}

		if cur == nil {
			// 两者没有关注关系
			relation.Link = dao.LinkForward
		} else {
			// 两者有过关注关系
			if (cur.UserAlpha == follower && cur.Link.IsForward()) ||
				(cur.UserBeta == follower && cur.Link.IsBackward()) ||
				(cur.Link == dao.LinkMutual) {
				// 无需重复关注
				return global.ErrAlreadyFollow
			} else {
				if (cur.UserAlpha == followee && cur.Link.IsForward()) ||
					(cur.UserBeta == followee && cur.Link.IsBackward()) {
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
		return xerror.Wrapf(err, "relation biz follow user transact failed").
			WithExtras("follower", follower, "followee", followee).WithCtx(ctx)
	}

	// set cache
	if err = infra.Dao().RelationCache.Follow(ctx, follower, relation); err != nil {
		xlog.Msg("relation biz set cache follow failed").
			Extra("relation", relation).
			Err(err).Errorx(ctx)
	}

	return nil
}

// follower取消对followee关注
func (b *RelationBiz) UserUnFollow(ctx context.Context, follower, followee int64) error {
	var (
		now      = time.Now().Unix()
		relation = &dao.Relation{
			UserAlpha: follower,
			UserBeta:  followee,
			Actime:    now,
			Amtime:    now,
		}
	)

	err := infra.Dao().DB().Transact(ctx, func(ctx context.Context) error {
		cur, err := infra.Dao().RelationDao.FindByAlphaBeta(ctx, follower, followee, true)
		if err != nil {
			if !errors.Is(err, xsql.ErrNoRecord) {
				return xerror.Wrapf(err, "dao find by alpha and beta failed")
			}
		}

		// 原本没有关注关系 无需处理
		if cur == nil || cur.Link == dao.LinkVacant {
			return nil
		}

		if (cur.UserAlpha == follower && cur.Link.IsForward()) ||
			(cur.UserBeta == follower && cur.Link.IsBackward()) {
			// 原本只是follower的单向关注，此时只需要将链接断开
			relation.Link = dao.LinkVacant
		} else if cur.Link.IsMutual() {
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

	if err = infra.Dao().RelationCache.UnFollow(ctx, follower, relation); err != nil {
		xlog.Msg("relation biz set cache unfollow failed").
			Extra("relation", relation).
			Err(err).Errorx(ctx)
	}

	return nil
}

// 获取用户的关注列表
func (b *RelationBiz) GetUserFollowingList(ctx context.Context, uid int64, offset int64, limit int) (
	[]int64, model.ListResult, error) {
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
//
// TODO Delay-tolerant
func (b *RelationBiz) GetUserFansList(ctx context.Context, uid int64, offset int64, limit int) (
	[]int64, model.ListResult, error) {
	var lr model.ListResult
	fans, next, more, err := infra.Dao().RelationDao.FindUidGotLinked(ctx, uid, offset, limit)
	if err != nil {
		return nil, lr, xerror.Wrapf(err, "relation biz failed to find uid got linked to").
			WithExtras("offset", offset, "limit", limit).WithCtx(ctx)
	}

	lr.NextOffset = next
	lr.HasMore = more
	return fans, lr, nil
}

// 获取用户关注数
func (b *RelationBiz) GetUserFollowingCount(ctx context.Context, uid int64) (int64, error) {
	cnt, err := infra.Dao().RelationDao.CountUidFollowings(ctx, uid)
	if err != nil {
		return 0, xerror.Wrapf(err, "relation biz failed to count user followings").WithCtx(ctx)
	}

	return cnt, nil
}

// 获取用户粉丝数
//
// Delay-tolerant
func (b *RelationBiz) GetUserFanCount(ctx context.Context, uid int64) (int64, error) {
	cnt, err := infra.Dao().RelationDao.CountUidFans(ctx, uid)
	if err != nil {
		return 0, xerror.Wrapf(err, "relation biz failed to count user fans").WithCtx(ctx)
	}

	return cnt, nil
}

// 检查uid是否关注了others
func (b *RelationBiz) BatchCheckUserFollowStatus(ctx context.Context, uid int64, others []int64) (map[int64]bool, error) {
	others = xslice.Uniq(others)
	others = xslice.Filter(others, func(_ int, v int64) bool { return v == uid })
	if len(others) == 0 {
		return nil, xerror.ErrArgs.Msg("no following relations")
	}

	cachePairs := make([]dao.UserPair, 0, len(others))
	for _, o := range others {
		var ua, ub int64
		if uid < o {
			ua = uid
			ub = o
		} else {
			ua = o
			ub = uid
		}
		cachePairs = append(cachePairs, dao.UserPair{
			UserA: ua,
			UserB: ub,
		})
	}

	var (
		daoResult           []*dao.RelationUser
		newlyFoundDaoResult []*dao.RelationUser
	)

	cacheResult, err := infra.Dao().RelationCache.BatchGetLinks(ctx, cachePairs)
	missingOthers := make([]int64, 0, len(others))
	existingOthers := make(map[int64]struct{}, len(others))
	if err == nil {
		for _, cl := range cacheResult {
			if cl.UserAlpha == uid {
				existingOthers[cl.UserBeta] = struct{}{}
			} else {
				existingOthers[cl.UserBeta] = struct{}{}
			}
		}
	}
	for _, o := range others {
		if _, ok := existingOthers[o]; !ok {
			missingOthers = append(missingOthers, o)
		}
	}

	if len(missingOthers) == 0 {
		daoResult = cacheResult
	} else {
		newlyFoundDaoResult, err = infra.Dao().RelationDao.BatchFindUidLinkTo(ctx, uid, missingOthers)
		if err != nil {
			return nil, xerror.Wrapf(err, "relation biz failed to find all followings").
				WithExtras("others", missingOthers).
				WithCtx(ctx)
		}

		daoResult = append(cacheResult, newlyFoundDaoResult...)
	}

	var followings = make([]int64, 0, len(others))
	for _, f := range daoResult {
		if f.UserAlpha == uid {
			followings = append(followings, f.UserBeta)
		} else {
			followings = append(followings, f.UserAlpha)
		}
	}

	res := make(map[int64]bool, len(others))
	fm := xslice.AsMap(followings)
	for _, o := range others {
		if _, ok := fm[o]; ok {
			res[o] = true
		}
	}

	// set cache
	if len(newlyFoundDaoResult) != 0 {
		concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
			Name: "relation.biz.batchcheckstatus.cache.batchset",
			Job: func(ctx context.Context) error {
				cacheData := make([]dao.CacheLink, 0, len(newlyFoundDaoResult))
				for _, r := range newlyFoundDaoResult {
					cacheData = append(cacheData, dao.CacheLink{
						Alpha: r.UserAlpha,
						Beta:  r.UserBeta,
						Link:  r.Link,
					})
				}
				err := infra.Dao().RelationCache.BatchSetLinks(ctx, cacheData)
				if err != nil {
					xlog.Msg("relation biz batch set link failed").Err(err).Errorx(ctx)
				}

				return nil
			},
		})
	}

	return res, nil
}

// 检查uid是否关注了other
func (b *RelationBiz) CheckUserFollowStatus(ctx context.Context, uid, other int64) (bool, error) {
	var (
		relLink      dao.LinkStatus
		checkForward bool
	)

	cacheData, err := infra.Dao().RelationCache.GetLink(ctx, uid, other)
	if err != nil {
		rel, err := infra.Dao().RelationDao.FindByAlphaBeta(ctx, uid, other, false)
		if err != nil {
			return false, xerror.Wrapf(err, "relation biz failed to find by alpha-beta").
				WithExtras("uid", uid, "other", other).
				WithCtx(ctx)
		}
		relLink = rel.Link
		if rel.UserAlpha == uid {
			checkForward = true
		}

		concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
			Name: "relation.biz.checkstatus.cache.set",
			Job: func(ctx context.Context) error {
				if err := infra.Dao().RelationCache.SetLink(ctx, uid, other, rel.Link); err != nil {
					xlog.Msg("relation biz set link cache failed").Err(err).Errorx(ctx)
				}

				return nil
			},
		})
	} else {
		// err == nil
		relLink = cacheData.Link
		if cacheData.UserAlpha == uid {
			checkForward = true
		}
	}

	if relLink.IsMutual() {
		return true, nil
	}

	if checkForward {
		return relLink.IsForward(), nil
	} else {
		// beta == uid
		return relLink.IsBackward(), nil
	}
}
