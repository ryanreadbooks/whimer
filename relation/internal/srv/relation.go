package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/relation/internal/biz"
	"github.com/ryanreadbooks/whimer/relation/internal/global"
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
