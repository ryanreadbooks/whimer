package handler

import (
	"net/http"

	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
	ptp "github.com/ryanreadbooks/whimer/passport/internal/model/trans/profile"
	"github.com/ryanreadbooks/whimer/passport/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ProfileMe(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		signedUser := model.CtxGetMeInfo(r.Context())
		if signedUser == nil {
			httpx.Error(w, global.ErrNotSignedIn)
			return
		}

		info, err := ctx.ProfileSvc.GetMe(r.Context(), signedUser.Uid)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, info)
	}
}

func ProfileUpdateMe(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		signedUser := model.CtxGetMeInfo(r.Context())
		if signedUser == nil {
			httpx.Error(w, global.ErrNotSignedIn)
			return
		}

		var req ptp.UpdateMeReq
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.Error(w, global.ErrArgs.Msg(err.Error()))
			return
		}

		if err := req.Validate(); err != nil {
			httpx.Error(w, err)
			return
		}

		if signedUser.Uid != req.Uid {
			httpx.Error(w, global.ErrPermDenied)
			return
		}

		me, err := ctx.ProfileSvc.UpdateMe(r.Context(), &req)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, me)
	}
}
