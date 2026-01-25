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

func (a *RelationAdapterImpl) CheckFollowed(ctx context.Context, uid, target int64) (bool, error) {
	resp, err := a.relationCli.CheckUserFollowed(ctx, &relationv1.CheckUserFollowedRequest{
		Uid:   uid,
		Other: target,
	})
	if err != nil {
		return false, xerror.Wrap(err)
	}
	return resp.GetFollowed(), nil
}

func (a *RelationAdapterImpl) GetFanCount(ctx context.Context, uid int64) (int64, error) {
	resp, err := a.relationCli.GetUserFanCount(ctx, &relationv1.GetUserFanCountRequest{
		Uid: uid,
	})
	if err != nil {
		return 0, xerror.Wrap(err)
	}
	return resp.GetCount(), nil
}

func (a *RelationAdapterImpl) GetFollowingCount(ctx context.Context, uid int64) (int64, error) {
	resp, err := a.relationCli.GetUserFollowingCount(ctx, &relationv1.GetUserFollowingCountRequest{
		Uid: uid,
	})
	if err != nil {
		return 0, xerror.Wrap(err)
	}
	return resp.GetCount(), nil
}

func (a *RelationAdapterImpl) PageGetFanList(ctx context.Context, uid int64, page, count int32) ([]int64, int64, error) {
	resp, err := a.relationCli.PageGetUserFanList(ctx, &relationv1.PageGetUserFanListRequest{
		Target: uid,
		Page:   page,
		Count:  count,
	})
	if err != nil {
		return nil, 0, xerror.Wrap(err)
	}
	return resp.GetFansId(), resp.GetTotal(), nil
}

func (a *RelationAdapterImpl) PageGetFollowingList(ctx context.Context, uid int64, page, count int32) ([]int64, int64, error) {
	resp, err := a.relationCli.PageGetUserFollowingList(ctx, &relationv1.PageGetUserFollowingListRequest{
		Target: uid,
		Page:   page,
		Count:  count,
	})
	if err != nil {
		return nil, 0, xerror.Wrap(err)
	}
	return resp.GetFollowingsId(), resp.GetTotal(), nil
}

func (a *RelationAdapterImpl) GetUserFollowingList(
	ctx context.Context, uid int64, offset int64, count int32,
) (*repository.FollowingListResult, error) {
	resp, err := a.relationCli.GetUserFollowingList(ctx, &relationv1.GetUserFollowingListRequest{
		Uid: uid,
		Cond: &relationv1.QueryCondition{
			Offset: offset,
			Count:  count,
		},
	})
	if err != nil {
		return nil, xerror.Wrap(err)
	}
	return &repository.FollowingListResult{
		Followings:  resp.GetFollowings(),
		FollowTimes: resp.GetFollowTimes(),
		HasMore:     resp.GetHasMore(),
		NextOffset:  resp.GetNextOffset(),
	}, nil
}
