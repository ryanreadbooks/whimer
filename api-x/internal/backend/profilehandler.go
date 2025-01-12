package backend

import (
	"net/http"
	"strconv"

	"github.com/ryanreadbooks/whimer/api-x/internal/backend/note"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/passport"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/profile"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/relation"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
	userv1 "github.com/ryanreadbooks/whimer/passport/sdk/user/v1"
	"github.com/zeromicro/go-zero/rest/httpx"

	relationv1 "github.com/ryanreadbooks/whimer/relation/sdk/v1"

	"golang.org/x/sync/errgroup"
)

// 获取用户的投稿数量、点赞数量等信息
func (h *Handler) GetProfileStat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			uid  = metadata.Uid(r.Context())
			stat struct {
				Posted     uint64 `json:"posted"`
				Fans       uint64 `json:"fans"`
				Followings uint64 `json:"followings"`
			}
		)

		eg, ctx := errgroup.WithContext(r.Context())

		// 用户投稿数量
		eg.Go(func() error {
			resp, err := note.NoteCreatorServer().GetPostedCount(ctx, &notev1.GetPostedCountRequest{
				Uid: uid,
			})
			if err != nil {
				return err
			}

			stat.Posted = resp.Count
			return nil
		})

		// 用户粉丝数量
		eg.Go(func() error {
			resp, err := relation.RelationServer().GetUserFanCount(ctx,
				&relationv1.GetUserFanCountRequest{
					Uid: uid,
				})
			if err != nil {
				return err
			}

			stat.Fans = resp.Count
			return nil
		})

		eg.Go(func() error {
			// 用户关注数量
			resp, err := relation.RelationServer().GetUserFollowingCount(ctx,
				&relationv1.GetUserFollowingCountRequest{
					Uid: uid,
				})
			if err != nil {
				return err
			}

			stat.Followings = resp.Count
			return nil
		})

		err := eg.Wait()
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, &stat)
	}
}

// 获取用户卡片信息
func (h *Handler) GetHoverProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx             = r.Context()
			uid             = metadata.Uid(ctx)
			isAuthedRequest = uid != 0
		)

		req, err := xhttp.ParseValidate[profile.HoverReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		eg, ctx := errgroup.WithContext(ctx)
		var (
			targetUser *userv1.UserInfo
		)

		// 基本信息
		eg.Go(func() error {
			return recovery.Do(func() error {
				res, err := passport.Userer().GetUser(ctx, &userv1.GetUserRequest{Uid: req.UserId})
				if err != nil {
					return err
				}

				targetUser = res.GetUser()
				return nil
			})
		})

		var (
			fansCount    uint64 = 0
			followsCount uint64 = 0
		)

		// 用户的粉丝，关注等信息
		eg.Go(func() error {
			return recovery.Do(func() error {
				// 粉丝数
				fanCntRes, err := relation.RelationServer().
					GetUserFanCount(ctx, &relationv1.GetUserFanCountRequest{
						Uid: req.UserId,
					})
				if err != nil {
					return err
				}

				fansCount = fanCntRes.GetCount()

				// 关注数
				followCntRes, err := relation.RelationServer().
					GetUserFollowingCount(ctx, &relationv1.GetUserFollowingCountRequest{
						Uid: req.UserId,
					})
				if err != nil {
					return err
				}

				followsCount = followCntRes.GetCount()

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
					followRes, _ := relation.RelationServer().CheckUserFollowed(ctx, &relationv1.CheckUserFollowedRequest{
						Uid:   uid,
						Other: req.UserId,
					})

					followed = followRes.GetFollowed()
					return nil
				})
			})
		}

		err = eg.Wait()
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// organize all result
		var res profile.HoverRes
		res.Relation.Status = profile.RelationNone
		res.BasicInfo.Nickname = targetUser.GetNickname()
		res.BasicInfo.StyleSign = targetUser.GetStyleSign()
		res.BasicInfo.Avatar = targetUser.GetAvatar()
		if !isAuthedRequest {
			// 非登录用户不展示准确的用户数据
			res.Interaction.Fans = profile.HideActualCount(fansCount)
			res.Interaction.Follows = profile.HideActualCount(followsCount)
		} else {
			res.Interaction.Fans = strconv.FormatUint(fansCount, 10)
			res.Interaction.Follows = strconv.FormatUint(followsCount, 10)
		}
		if followed {
			res.Relation.Status = profile.RelationFollowing
		}
		xhttp.OkJson(w, res)
	}
}
