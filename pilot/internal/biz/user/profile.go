package user

import (
	"context"
	"strconv"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/recovery"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/user/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"

	"golang.org/x/sync/errgroup"
)

// 获取用户卡片信息
func (b *Biz) GetHoverProfile(ctx context.Context, targetUid int64) (*model.HoverInfo, error) {
	var (
		uid             = metadata.Uid(ctx)
		isAuthedRequest = uid != 0
	)

	eg, ctx := errgroup.WithContext(ctx)
	var (
		targetUser *userv1.UserInfo
	)

	// 基本信息
	eg.Go(func() error {
		return recovery.Do(func() error {
			res, err := dep.Userer().GetUser(ctx, &userv1.GetUserRequest{Uid: targetUid})
			if err != nil {
				return err
			}

			targetUser = res.GetUser()
			return nil
		})
	})

	var (
		fansCount    int64 = 0
		followsCount int64 = 0
	)

	// 用户的粉丝，关注等信息
	eg.Go(func() error {
		return recovery.Do(func() error {
			// 粉丝数
			fanCntRes, err := dep.RelationServer().
				GetUserFanCount(ctx, &relationv1.GetUserFanCountRequest{
					Uid: targetUid,
				})
			if err != nil {
				return err
			}

			fansCount = fanCntRes.GetCount()

			// 关注数
			followCntRes, err := dep.RelationServer().
				GetUserFollowingCount(ctx, &relationv1.GetUserFollowingCountRequest{
					Uid: targetUid,
				})
			if err != nil {
				return err
			}

			followsCount = followCntRes.GetCount()

			return nil
		})
	})

	// 用户最近发布的笔记信息
	var postAssets = make([]model.PostAsset, 0, 3)
	eg.Go(func() error {
		return recovery.Do(func() error {
			resp, err := dep.NoteFeedServer().GetUserRecentPost(ctx, &notev1.GetUserRecentPostRequest{
				Uid:   targetUid,
				Count: 3,
			})
			if err != nil {
				return err
			}

			for _, item := range resp.Items {
				// 此处只需要封面
				if len(item.Images) > 0 {
					postAssets = append(postAssets, model.PostAsset{
						Url:    item.Images[0].Url,
						UrlPrv: item.Images[0].UrlPrv,
						Type:   int(item.Images[0].Type),
					})
				}
			}

			return nil
		})
	})

	var (
		followed bool
	)
	if isAuthedRequest {
		// 获取请求用户和目标用户的关注关系
		eg.Go(func() error {
			return recovery.Do(func() error {
				followRes, _ := dep.RelationServer().CheckUserFollowed(ctx, &relationv1.CheckUserFollowedRequest{
					Uid:   uid,
					Other: targetUid,
				})

				followed = followRes.GetFollowed()
				return nil
			})
		})
	}

	err := eg.Wait()
	if err != nil {
		return nil, err
	}

	// organize all result
	var res model.HoverInfo
	res.Relation.Status = model.RelationNone
	res.BasicInfo.Nickname = targetUser.GetNickname()
	res.BasicInfo.StyleSign = targetUser.GetStyleSign()
	res.BasicInfo.Avatar = targetUser.GetAvatar()
	if !isAuthedRequest {
		// 非登录用户不展示准确的用户数据
		res.Interaction.Fans = model.HideActualCount(fansCount)
		res.Interaction.Follows = model.HideActualCount(followsCount)
	} else {
		res.Interaction.Fans = strconv.FormatInt(fansCount, 10)
		res.Interaction.Follows = strconv.FormatInt(followsCount, 10)
	}
	if followed {
		res.Relation.Status = model.RelationFollowing
	}
	res.RecentPosts = postAssets

	return &res, nil
}
