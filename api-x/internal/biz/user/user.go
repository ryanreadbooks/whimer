package user

import (
	"context"

	"github.com/ryanreadbooks/whimer/api-x/internal/biz/user/model"
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"
	"golang.org/x/sync/errgroup"
)

type UserBiz struct{}

func NewUserBiz(c *config.Config) *UserBiz {
	return &UserBiz{}
}

func (b *UserBiz) ListUsers(ctx context.Context, uids []int64) (map[string]*userv1.UserInfo, error) {
	resp, err := infra.Userer().BatchGetUser(ctx, &userv1.BatchGetUserRequest{
		Uids: uids,
	})
	if err != nil {
		return nil, err
	}

	return resp.GetUsers(), nil
}

func (b *UserBiz) GetUser(ctx context.Context, uid int64) (*userv1.UserInfo, error) {
	resp, err := infra.Userer().GetUser(ctx, &userv1.GetUserRequest{
		Uid: uid,
	})
	if err != nil {
		return nil, err
	}

	return resp.GetUser(), nil
}

func (b *UserBiz) batchGetFollowingStatus(ctx context.Context, uid int64, targets []int64, output []*model.UserWithFollowStatus) (map[int64]bool, error) {
	resp, err := infra.RelationServer().BatchCheckUserFollowed(ctx,
		&relationv1.BatchCheckUserFollowedRequest{
			Uid:     uid,
			Targets: targets,
		})
	if err != nil {
		return nil, xerror.Wrapf(err, "batch check user followed failed")
	}

	gotStatus := resp.GetStatus()

	// 填充out中的following status
	for _, item := range output {
		if item == nil || item.User == nil {
			continue
		}

		targetUid := item.User.Uid // target
		if followed, ok := gotStatus[targetUid]; ok && followed {
			item.Relation = model.RelationFollowing
		} else {
			item.Relation = model.RelationNone
		}
	}

	return gotStatus, nil
}

func (b *UserBiz) getUserWithFollowStatus(ctx context.Context, targetUids []int64) ([]*model.UserWithFollowStatus, error) {
	var (
		uid             = metadata.Uid(ctx)
		isAuthedRequest = uid != 0
	)

	var (
		result = make([]*model.UserWithFollowStatus, len(targetUids))
	)

	if len(targetUids) > 0 {
		userResp, err := infra.Userer().BatchGetUserV2(ctx, &userv1.BatchGetUserV2Request{
			Uids: targetUids,
		})
		if err != nil {
			return nil, err
		}

		for idx, targetUid := range targetUids {
			user := userResp.Users[targetUid]
			result[idx] = &model.UserWithFollowStatus{
				User:     user,
				Relation: model.RelationNone,
			}
		}
	}

	if isAuthedRequest {
		_, err := b.batchGetFollowingStatus(ctx, uid, targetUids, result)
		if err != nil {
			xlog.Msg("user biz batch get following status failed").Err(err).Errorx(ctx)
		}
	}

	return result, nil
}

// 获取用户粉丝列表
func (b *UserBiz) ListUserFans(ctx context.Context, targetUid int64, page, count int32) ([]*model.UserWithFollowStatus, int64, error) {
	resp, err := infra.RelationServer().PageGetUserFanList(ctx,
		&relationv1.PageGetUserFanListRequest{
			Target: targetUid,
			Page:   page,
			Count:  count,
		})

	if err != nil {
		return nil, 0, err
	}

	var (
		fansId       = resp.GetFansId()
		total  int64 = resp.Total
	)

	result, err := b.getUserWithFollowStatus(ctx, fansId)
	if err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

// 获取用户关注列表
func (b *UserBiz) ListUserFollowings(ctx context.Context, targetUid int64, page, count int32) ([]*model.UserWithFollowStatus, int64, error) {
	resp, err := infra.RelationServer().PageGetUserFollowingList(ctx,
		&relationv1.PageGetUserFollowingListRequest{
			Target: targetUid,
			Page:   page,
			Count:  count,
		})

	if err != nil {
		return nil, 0, err
	}

	var (
		followingsId       = resp.GetFollowingsId()
		total        int64 = resp.Total
	)

	result, err := b.getUserWithFollowStatus(ctx, followingsId)
	if err != nil {
		return nil, 0, err
	}

	return result, total, nil
}

// 获取用户的投稿数量、点赞数量等信息
func (b *UserBiz) GetUserStat(ctx context.Context, targetUid int64) (*model.UserStat, error) {
	var (
		reqUid = metadata.Uid(ctx)
		stat   model.UserStat
	)

	eg, ctx := errgroup.WithContext(ctx)
	// 用户投稿数量
	eg.Go(func() error {
		var (
			err error
			cnt int64
		)

		if reqUid == targetUid {
			var resp *notev1.GetPostedCountResponse
			resp, err = infra.NoteCreatorServer().GetPostedCount(ctx, &notev1.GetPostedCountRequest{
				Uid: reqUid,
			})
			if resp != nil {
				cnt = resp.Count
			}
		} else {
			var resp *notev1.GetPublicPostedCountResponse
			resp, err = infra.NoteFeedServer().GetPublicPostedCount(ctx, &notev1.GetPublicPostedCountRequest{
				Uid: targetUid,
			})
			if resp != nil {
				cnt = resp.Count
			}
		}
		if err != nil {
			return err
		}

		stat.Posted = cnt
		return nil
	})

	// 用户粉丝数量
	eg.Go(func() error {
		resp, err := infra.RelationServer().GetUserFanCount(ctx,
			&relationv1.GetUserFanCountRequest{
				Uid: targetUid,
			})
		if err != nil {
			return err
		}

		stat.Fans = resp.Count
		return nil
	})

	eg.Go(func() error {
		// 用户关注数量
		resp, err := infra.RelationServer().GetUserFollowingCount(ctx,
			&relationv1.GetUserFollowingCountRequest{
				Uid: targetUid,
			})
		if err != nil {
			return err
		}

		stat.Followings = resp.Count
		return nil
	})

	err := eg.Wait()
	if err != nil {
		return &stat, err
	}

	return &stat, nil
}
