package relation

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/relation/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/relation/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/relation/convert"
	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"
)

type RelationAdapterImpl struct {
	relationCli relationv1.RelationServiceClient
}

var _ repository.RelationAdapter = &RelationAdapterImpl{}

func NewRelationAdapterImpl(c relationv1.RelationServiceClient) *RelationAdapterImpl {
	return &RelationAdapterImpl{
		relationCli: c,
	}
}

func (a *RelationAdapterImpl) BatchGetFollowStatus(
	ctx context.Context, uid int64, targets []int64,
) (map[int64]bool, error) {
	resp, err := a.relationCli.BatchCheckUserFollowed(ctx,
		&relationv1.BatchCheckUserFollowedRequest{
			Uid:     uid,
			Targets: targets,
		})
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	return resp.Status, nil
}

func (a *RelationAdapterImpl) FollowUser(
	ctx context.Context, follower, followee int64, action vo.FollowAction,
) error {
	_, err := a.relationCli.FollowUser(ctx,
		&relationv1.FollowUserRequest{
			Follower: follower,
			Followee: followee,
			Action:   convert.VoFollowActionToPb(action),
		})
	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}

func (a *RelationAdapterImpl) UpdateSettings(
	ctx context.Context, uid int64, showFans, showFollows bool,
) error {
	_, err := a.relationCli.UpdateUserSettings(ctx,
		&relationv1.UpdateUserSettingsRequest{
			TargetUid:      uid,
			ShowFanList:    showFans,
			ShowFollowList: showFollows,
		})
	if err != nil {
		return xerror.Wrap(err)
	}

	return nil
}
