package user

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/app/user/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/user/vo"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

// 分页获取用户粉丝列表
func (s *Service) ListUserFans(ctx context.Context, targetUid int64, page, count int32) (*dto.FanOrFollowingListResult, error) {
	fansId, total, err := s.relationAdapter.PageGetFanList(ctx, targetUid, page, count)
	if err != nil {
		return nil, xerror.Wrapf(err, "page get fan list failed").WithCtx(ctx)
	}

	usersWithRelation, err := s.getUserWithFollowStatus(ctx, fansId)
	if err != nil {
		return nil, xerror.Wrapf(err, "get user with follow status failed").WithCtx(ctx)
	}

	items := make([]*dto.UserWithRelation, 0, len(usersWithRelation))
	for _, u := range usersWithRelation {
		items = append(items, dto.ConvertVoUserWithRelationToDto(u))
	}

	return &dto.FanOrFollowingListResult{Items: items, Total: total}, nil
}

// 分页获取用户关注列表
func (s *Service) ListUserFollowings(ctx context.Context, targetUid int64, page, count int32) (*dto.FanOrFollowingListResult, error) {
	followingsId, total, err := s.relationAdapter.PageGetFollowingList(ctx, targetUid, page, count)
	if err != nil {
		return nil, xerror.Wrapf(err, "page get following list failed").WithCtx(ctx)
	}

	usersWithRelation, err := s.getUserWithFollowStatus(ctx, followingsId)
	if err != nil {
		return nil, xerror.Wrapf(err, "get user with follow status failed").WithCtx(ctx)
	}

	items := make([]*dto.UserWithRelation, 0, len(usersWithRelation))
	for _, u := range usersWithRelation {
		items = append(items, dto.ConvertVoUserWithRelationToDto(u))
	}

	return &dto.FanOrFollowingListResult{Items: items, Total: total}, nil
}

// 批量获取关注状态并填充到结果中
func (s *Service) batchGetFollowingStatus(
	ctx context.Context,
	uid int64,
	targets []int64,
	output []*vo.UserWithRelation,
) (map[int64]bool, error) {
	statuses, err := s.relationAdapter.BatchGetFollowStatus(ctx, uid, targets)
	if err != nil {
		return nil, xerror.Wrapf(err, "batch check user followed failed")
	}

	for _, item := range output {
		if item == nil || item.User == nil {
			continue
		}
		if followed, ok := statuses[item.User.Uid]; ok && followed {
			item.Relation = vo.RelationFollowing
		} else {
			item.Relation = vo.RelationNone
		}
	}

	return statuses, nil
}

// 获取用户及其关注状态
func (s *Service) getUserWithFollowStatus(ctx context.Context, targetUids []int64) ([]*vo.UserWithRelation, error) {
	uid := metadata.Uid(ctx)
	isAuthedRequest := uid != 0

	result := make([]*vo.UserWithRelation, len(targetUids))
	if len(targetUids) > 0 {
		users, err := s.userAdapter.BatchGetUser(ctx, targetUids)
		if err != nil {
			return nil, err
		}

		for idx, targetUid := range targetUids {
			result[idx] = &vo.UserWithRelation{
				User:     users[targetUid],
				Relation: vo.RelationNone,
			}
		}
	}

	if isAuthedRequest {
		_, err := s.batchGetFollowingStatus(ctx, uid, targetUids, result)
		if err != nil {
			xlog.Msg("batch get following status failed").Err(err).Errorx(ctx)
		}
	}

	return result, nil
}
