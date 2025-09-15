package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/relation/internal/biz"
	"github.com/ryanreadbooks/whimer/relation/internal/global"
	"github.com/ryanreadbooks/whimer/relation/internal/infra"
	"github.com/ryanreadbooks/whimer/relation/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/relation/internal/model"

	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xconv"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type RelationSrv struct {
	Ctx *Service

	relationBiz *biz.RelationBiz
}

func NewRelationSrv(p *Service, biz biz.Biz) *RelationSrv {
	s := &RelationSrv{
		Ctx:         p,
		relationBiz: biz.Relation,
	}

	return s
}

const (
	followUserLockExpireSec = 10
)

func fmtFollowUserLockKey(follower, followed int64) string {
	// relation:srv:follow:lock:uid1>uid2
	return "relation:srv:follow:lock:" + xconv.FormatInt(follower) + ">" + xconv.FormatInt(followed)
}

// follower关注followed
func (s *RelationSrv) FollowUser(ctx context.Context, follower, followed int64) error {
	var (
		uid = metadata.Uid(ctx)
	)

	if uid != follower {
		return global.ErrPermDenied
	}

	if uid == followed {
		return global.ErrFollowSelf
	}

	if err := s.checkUserExistence(ctx, followed); err != nil {
		return xerror.Wrapf(err, "check user existence failed")
	}

	lock := redis.NewRedisLock(infra.Cache(), fmtFollowUserLockKey(follower, followed))
	lock.SetExpire(followUserLockExpireSec)
	hasLock, err := lock.AcquireCtx(ctx)
	if err != nil {
		return xerror.Wrapf(err, "follow user failed to acquire lock")
	}
	if !hasLock {
		return xerror.Wrap(global.ErrLockNotHeld)
	}
	defer lock.ReleaseCtx(ctx)

	curCnt, err := s.relationBiz.GetUserFollowingCount(ctx, uid)
	if err != nil {
		return xerror.Wrapf(err, "relation service check follow count failed").WithCtx(ctx)
	}
	if curCnt >= global.MaxFollowAllowed {
		return global.ErrFollowReachMaxCount
	}

	err = s.relationBiz.UserFollow(ctx, uid, followed)
	if err != nil {
		return xerror.Wrapf(err, "relation service user follow failed").WithCtx(ctx)
	}

	return nil
}

// follower取关unfollowed
func (s *RelationSrv) UnfollowUser(ctx context.Context, follower, unfollowed int64) error {
	var (
		uid = metadata.Uid(ctx)
	)

	if uid != follower {
		return global.ErrPermDenied
	}

	if uid == unfollowed {
		return global.ErrUnFollowSelf
	}

	if err := s.checkUserExistence(ctx, unfollowed); err != nil {
		return xerror.Wrapf(err, "check user existence failed")
	}

	lock := redis.NewRedisLock(infra.Cache(), fmtFollowUserLockKey(follower, unfollowed))
	lock.SetExpire(followUserLockExpireSec)
	hasLock, err := lock.AcquireCtx(ctx)
	if err != nil {
		return xerror.Wrapf(err, "follow user failed to acquire lock")
	}
	if !hasLock {
		return xerror.Wrap(global.ErrLockNotHeld)
	}
	defer lock.ReleaseCtx(ctx)

	err = s.relationBiz.UserUnFollow(ctx, uid, unfollowed)
	if err != nil {
		return xerror.Wrapf(err, "relation service unfollow user failed")
	}

	return nil
}

// 获取粉丝列表
func (s *RelationSrv) GetUserFanList(ctx context.Context, who int64, offset int64, cnt int) (
	fans []int64, result model.ListResult, err error) {

	var (
		uid = metadata.Uid(ctx)
	)

	if uid != who {
		err = xerror.Wrap(global.ErrNotAllowedGetFanList)
		return
	}

	fans, result, err = s.relationBiz.GetUserFansList(ctx, who, offset, cnt)
	if err != nil {
		err = xerror.Wrapf(err, "relation service get user fans list failed")
		return
	}

	return
}

// 获取关注列表
func (s *RelationSrv) GetUserFollowingList(ctx context.Context, who int64, offset int64, cnt int) (
	followings []int64, result model.ListResult, err error) {

	var (
		uid = metadata.Uid(ctx)
	)

	if uid != who {
		err = xerror.Wrap(global.ErrNotAllowedGetFollowingList)
		return
	}

	followings, result, err = s.relationBiz.GetUserFollowingList(ctx, who, offset, cnt)
	if err != nil {
		err = xerror.Wrapf(err, "relation service get user following list failed")
		return
	}

	return
}

// 获取用户粉丝数
func (s *RelationSrv) GetUserFanCount(ctx context.Context, uid int64) (int64, error) {
	return s.relationBiz.GetUserFanCount(ctx, uid)
}

// 获取用户关注数
func (s *RelationSrv) GetUserFollowingCount(ctx context.Context, uid int64) (int64, error) {
	return s.relationBiz.GetUserFollowingCount(ctx, uid)
}

// 检查用户是否关注过某些用户
func (s *RelationSrv) BatchCheckUserFollowStatus(ctx context.Context, uid int64, targets []int64) (map[int64]bool, error) {
	return s.relationBiz.BatchCheckUserFollowStatus(ctx, uid, targets)
}

func (s *RelationSrv) CheckUserFollowStatus(ctx context.Context, uid, other int64) (bool, error) {
	return s.relationBiz.CheckUserFollowStatus(ctx, uid, other)
}

func (s *RelationSrv) hasUser(ctx context.Context, uid int64) (bool, error) {
	r, err := dep.Userer().HasUser(ctx, &userv1.HasUserRequest{Uid: uid})
	if err != nil {
		return false, xerror.Wrapf(err, "relation check user failed").WithCtx(ctx).WithExtra("uid", uid)
	}

	return r.Has, nil
}

func (s *RelationSrv) checkUserExistence(ctx context.Context, uid int64) error {
	hasUser, err := s.hasUser(ctx, uid)
	if err != nil {
		return xerror.Wrap(err).WithCtx(ctx)
	}

	if !hasUser {
		return xerror.Wrap(global.ErrUserNotFound).WithCtx(ctx)
	}

	return nil
}
