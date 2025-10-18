package user

import (
	"context"
	"sort"
	"strings"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/user/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/cache/recentcontact"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"

	"golang.org/x/sync/errgroup"
)

type Biz struct {
	recentContact *recentcontact.Store
}

func NewUserBiz(c *config.Config) *Biz {
	return &Biz{
		recentContact: recentcontact.New(infra.Cache()),
	}
}

func (b *Biz) ListUsers(ctx context.Context, uids []int64) (map[string]*userv1.UserInfo, error) {
	resp, err := dep.Userer().BatchGetUser(ctx, &userv1.BatchGetUserRequest{
		Uids: uids,
	})
	if err != nil {
		return nil, err
	}

	return resp.GetUsers(), nil
}

func (b *Biz) ListUsersV2(ctx context.Context, uids []int64) (map[int64]*userv1.UserInfo, error) {
	resp, err := dep.Userer().BatchGetUserV2(ctx, &userv1.BatchGetUserV2Request{
		Uids: uids,
	})
	if err != nil {
		return nil, err
	}

	return resp.GetUsers(), nil
}

func (b *Biz) GetUser(ctx context.Context, uid int64) (*userv1.UserInfo, error) {
	resp, err := dep.Userer().GetUser(ctx, &userv1.GetUserRequest{
		Uid: uid,
	})
	if err != nil {
		return nil, err
	}

	return resp.GetUser(), nil
}

func (b *Biz) batchGetFollowingStatus(ctx context.Context,
	uid int64,
	targets []int64, output []*model.UserWithFollowStatus) (map[int64]bool, error) {

	resp, err := dep.RelationServer().BatchCheckUserFollowed(ctx,
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

func (b *Biz) getUserWithFollowStatus(ctx context.Context, targetUids []int64) (
	[]*model.UserWithFollowStatus, error) {
	var (
		uid             = metadata.Uid(ctx)
		isAuthedRequest = uid != 0
	)

	var (
		result = make([]*model.UserWithFollowStatus, len(targetUids))
	)

	if len(targetUids) > 0 {
		userResp, err := dep.Userer().BatchGetUserV2(ctx, &userv1.BatchGetUserV2Request{
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

// 分页获取用户粉丝列表
func (b *Biz) ListUserFans(ctx context.Context, targetUid int64, page, count int32) (
	[]*model.UserWithFollowStatus, int64, error) {

	resp, err := dep.RelationServer().PageGetUserFanList(ctx,
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

// 分页获取用户关注列表
func (b *Biz) ListUserFollowings(ctx context.Context, targetUid int64, page, count int32) (
	[]*model.UserWithFollowStatus, int64, error) {

	resp, err := dep.RelationServer().PageGetUserFollowingList(ctx,
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
func (b *Biz) GetUserStat(ctx context.Context, targetUid int64) (*model.UserStat, error) {
	var (
		reqUid = metadata.Uid(ctx)
		stat   model.UserStat
	)

	eg, ctx := errgroup.WithContext(ctx)
	// 用户投稿数量
	eg.Go(func() error {
		return recovery.Do(func() error {
			var (
				err error
				cnt int64
			)

			if reqUid == targetUid {
				var resp *notev1.GetPostedCountResponse
				resp, err = dep.NoteCreatorServer().GetPostedCount(ctx, &notev1.GetPostedCountRequest{
					Uid: reqUid,
				})
				if resp != nil {
					cnt = resp.Count
				}
			} else {
				var resp *notev1.GetPublicPostedCountResponse
				resp, err = dep.NoteFeedServer().GetPublicPostedCount(ctx, &notev1.GetPublicPostedCountRequest{
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
	})

	// 用户粉丝数量
	eg.Go(func() error {
		return recovery.Do(func() error {
			resp, err := dep.RelationServer().GetUserFanCount(ctx,
				&relationv1.GetUserFanCountRequest{
					Uid: targetUid,
				})
			if err != nil {
				return err
			}

			stat.Fans = resp.Count
			return nil
		})
	})

	eg.Go(func() error {
		return recovery.Do(func() error {
			// 用户关注数量
			resp, err := dep.RelationServer().GetUserFollowingCount(ctx,
				&relationv1.GetUserFollowingCountRequest{
					Uid: targetUid,
				})
			if err != nil {
				return err
			}

			stat.Followings = resp.Count
			return nil
		})
	})

	err := eg.Wait()
	if err != nil {
		return &stat, err
	}

	return &stat, nil
}

type UidAndTime struct {
	Uid  int64
	Time int64
}

type followingUser struct {
	*model.User
	followTime int64
}

// 按照nickname获取关注的用户
//
// 由于是不同服务存储关注关系和用户信息 所以此种方法可能在数据量大的时候很慢
//
// 这里只提供获取关注的方法不提供获取粉丝的方法 因为如果按照这种方式获取粉丝数量的话 在粉丝量庞大时会导致性能问题;
// 由于限制了关注的人数 所以暴力获取的方式应该能接受
func (b *Biz) BrutalListFollowingsByName(ctx context.Context, uid int64, target string) ([]*model.User, error) {
	// 全量拿关注列表
	var (
		offset int64 = 0
		count  int32 = 250
	)

	followings := make([]UidAndTime, 0, 128)
	for {
		tmpResp, err := dep.RelationServer().GetUserFollowingList(ctx,
			&relationv1.GetUserFollowingListRequest{
				Uid: uid,
				Cond: &relationv1.QueryCondition{
					Offset: offset,
					Count:  count,
				},
			})
		if err != nil {
			return nil, xerror.Wrapf(err, "remote relation server get user following list failed")
		}
		if len(tmpResp.Followings) == 0 {
			break
		}

		for idx := range tmpResp.Followings {
			followings = append(followings, UidAndTime{
				Uid:  tmpResp.Followings[idx],
				Time: tmpResp.FollowTimes[idx],
			})
		}

		if tmpResp.HasMore {
			offset = tmpResp.NextOffset
		} else {
			break
		}
	}

	if len(followings) == 0 {
		return []*model.User{}, nil
	}

	followingsMap := xslice.MakeMap(followings, func(v UidAndTime) int64 {
		return v.Uid
	})

	uids := xmap.Keys(followingsMap)
	users, err := dep.Userer().BatchGetUserV2(ctx, &userv1.BatchGetUserV2Request{
		Uids: uids,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "remote user server batch get user failed")
	}

	// 本地筛选nickname
	resultUsers := make([]*model.User, 0, len(users.Users))
	for _, user := range users.Users {
		if strings.Contains(user.Nickname, target) {
			resultUsers = append(resultUsers, user)
		}
	}

	resultUsersMap := xslice.MakeMap(resultUsers, func(v *model.User) int64 {
		return v.Uid
	})

	followingUsers := make([]*followingUser, 0, len(resultUsersMap))
	for _, user := range resultUsersMap {
		followingUsers = append(followingUsers, &followingUser{
			User:       user,
			followTime: followingsMap[user.Uid].Time,
		})
	}

	// 按照关注时间排序
	sort.Slice(followingUsers, func(i, j int) bool {
		return followingUsers[i].followTime > followingUsers[j].followTime
	})

	results := make([]*model.User, 0, len(followingUsers))
	for _, user := range followingUsers {
		results = append(results, user.User)
	}

	return results, nil
}
