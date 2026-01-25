package relation

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/relation/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/relation/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/relation/vo"
)

type Service struct {
	relationAdapter repository.RelationAdapter
}

func NewService(relationAdapter repository.RelationAdapter) *Service {
	return &Service{
		relationAdapter: relationAdapter,
	}
}

func (s *Service) FollowOrUnfollow(ctx context.Context, uid int64, cmd *dto.FollowCommand) error {
	err := s.relationAdapter.FollowUser(ctx, uid, cmd.Target, vo.FollowAction(cmd.Action))
	if err != nil {
		return xerror.Wrapf(err, "relation adapter follow user failed").
			WithExtras("uid", uid, "target", cmd.Target, "action", cmd.Action).WithCtx(ctx)
	}

	return nil
}

func (s *Service) CheckFollowing(ctx context.Context, uid, target int64) (bool, error) {
	statuses, err := s.relationAdapter.BatchGetFollowStatus(ctx, uid, []int64{target})
	if err != nil {
		return false, xerror.Wrapf(err, "relation adapter batch get follow status failed").
			WithExtras("uid", uid, "target", target).WithCtx(ctx)
	}

	return statuses[target], nil
}

func (s *Service) UpdateSettings(ctx context.Context, uid int64, cmd *dto.UpdateSettingsCommand) error {
	err := s.relationAdapter.UpdateSettings(ctx, uid, cmd.ShowFans, cmd.ShowFollows)
	if err != nil {
		return xerror.Wrapf(err, "relation adapter update settings failed").
			WithExtras("uid", uid).WithCtx(ctx)
	}

	return nil
}
