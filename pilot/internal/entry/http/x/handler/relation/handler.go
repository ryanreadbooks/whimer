package relation

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	bizrelation "github.com/ryanreadbooks/whimer/pilot/internal/biz/relation"
	bizmodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/relation/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct {
	relationBiz *bizrelation.Biz
}

func NewHandler(c *config.Config, bizz *biz.Biz) *Handler {
	return &Handler{
		relationBiz: bizz.RelationBiz,
	}
}

func (h *Handler) UserFollowAction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		req, err := xhttp.ParseValidate[bizmodel.FollowReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// 关注或者取消关注
		err = h.relationBiz.FollowOrUnfollow(ctx, uid, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}

// 检查是否关注了某个用户
func (h *Handler) GetIsFollowing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[bizmodel.GetIsFollowingReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		uid := metadata.Uid(ctx)
		if uid == 0 {
			// 未登录
			xhttp.OkJson(w, false)
			return
		}

		followed, err := h.relationBiz.CheckUserFollows(ctx, uid, req.Uid)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, followed)
	}
}

// 用户关注设置
func (h *Handler) UpdateSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[bizmodel.UpdateSettingReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		err = h.relationBiz.UpdateRelationSettings(ctx, uid, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}
