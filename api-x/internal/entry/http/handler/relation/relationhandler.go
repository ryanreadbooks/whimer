package relation

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	relationv1 "github.com/ryanreadbooks/whimer/relation/api/v1"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct{}

func NewHandler(c *config.Config) *Handler {
	return &Handler{}
}

func (h *Handler) UserFollowAction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		req, err := xhttp.ParseValidate[FollowReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// 关注或者取消关注
		resp, err := infra.RelationServer().FollowUser(ctx, &relationv1.FollowUserRequest{
			Follower: uid,
			Followee: req.Target,
			Action:   relationv1.FollowUserRequest_Action(req.Action),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		_ = resp
		xhttp.OkJson(w, nil)
	}
}

// 检查是否关注了某个用户
func (h *Handler) GetIsFollowing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[GetIsFollowingReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := infra.RelationServer().BatchCheckUserFollowed(ctx, &relationv1.BatchCheckUserFollowedRequest{
			Uid:     metadata.Uid(ctx),
			Targets: []int64{req.UserId},
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp.GetStatus()[req.UserId])
	}
}

// 获取粉丝列表
func (h *Handler) GetFansList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}

// 获取关注列表
func (h *Handler) GetFollowingsList() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
