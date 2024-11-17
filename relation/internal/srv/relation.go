package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/relation/internal/biz"
	"github.com/ryanreadbooks/whimer/relation/internal/global"
	"github.com/ryanreadbooks/whimer/relation/internal/model"
)

type RelationSrv struct {
	Ctx *Service

	relationBiz biz.RelationBiz
}

func NewRelationSrv(p *Service, biz biz.Biz) *RelationSrv {
	s := &RelationSrv{
		Ctx:         p,
		relationBiz: biz.Relation,
	}

	return s
}

func (s *RelationSrv) FollowUser(ctx context.Context, follower, followed uint64) error {
	var (
		uid = metadata.Uid(ctx)
	)

	if uid != follower {
		return global.ErrPermDenied
	}

	if uid == followed {
		return global.ErrFollowSelf
	}

	err := s.relationBiz.UserFollow(ctx, uid, followed)
	if err != nil {
		return xerror.Wrapf(err, "relation service user follow failed")
	}

	return nil
}

func (s *RelationSrv) UnfollowUser(ctx context.Context, follower, unfollowed uint64) error {
	var (
		uid = metadata.Uid(ctx)
	)

	if uid != follower {
		return global.ErrPermDenied
	}

	if uid == unfollowed {
		return global.ErrUnFollowSelf
	}

	err := s.relationBiz.UserUnFollow(ctx, uid, unfollowed)
	if err != nil {
		return xerror.Wrapf(err, "relation service unfollow user failed")
	}

	return nil
}

func (s *RelationSrv) GetUserFanList(ctx context.Context, who, offset uint64, cnt int) (fans []uint64, result model.ListResult, err error) {
	var (
		uid = metadata.Uid(ctx)
	)

	if uid != who {
		err = global.ErrNotAllowedGetFanList
		return
	}

	fans, result, err = s.relationBiz.GetUserFansList(ctx, who, offset, cnt)
	if err != nil {
		err = xerror.Wrapf(err, "relation service get user fans list failed")
		return
	}

	return
}

func (s *RelationSrv) GetUserFollowingList(ctx context.Context, who, offset uint64, cnt int) (followings []uint64, result model.ListResult, err error) {
	var (
		uid = metadata.Uid(ctx)
	)

	if uid != who {
		err = global.ErrNotAllowedGetFollowingList
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
func (s *RelationSrv) GetUserFanCount(ctx context.Context, uid uint64) (uint64, error) {
	return s.relationBiz.GetUserFanCount(ctx, uid)
}

// 获取用户关注数
func (s *RelationSrv) GetUserFollowingCount(ctx context.Context, uid uint64) (uint64, error) {
	return s.relationBiz.GetUserFollowingCount(ctx, uid)
}
