package relation

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/app"
	apprelation "github.com/ryanreadbooks/whimer/pilot/internal/app/relation"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/relation/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct {
	relationApp *apprelation.Service
}

func NewHandler(c *config.Config, manager *app.Manager) *Handler {
	return &Handler{
		relationApp: manager.RelationApp,
	}
}

func (h *Handler) UserFollowAction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.FollowCommand](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		uid := metadata.Uid(ctx)

		err = h.relationApp.FollowOrUnfollow(ctx, uid, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}

func (h *Handler) GetIsFollowing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.CheckFollowingQuery](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		uid := metadata.Uid(ctx)
		if uid == 0 {
			xhttp.OkJson(w, false)
			return
		}

		followed, err := h.relationApp.CheckFollowing(ctx, uid, req.Uid)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, followed)
	}
}

func (h *Handler) UpdateSettings() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.UpdateSettingsCommand](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		uid := metadata.Uid(ctx)

		err = h.relationApp.UpdateSettings(ctx, uid, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, nil)
	}
}
