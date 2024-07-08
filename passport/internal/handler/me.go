package handler

import (
	"net/http"
	"net/url"

	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
	"github.com/ryanreadbooks/whimer/passport/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func PassportMe(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(model.WhimerSessId)
		if err != nil || len(cookie.Value) == 0 {
			httpx.Error(w, global.ErrNotSignedIn)
			return
		}

		cv, err := url.PathUnescape(cookie.Value)
		if err != nil {
			httpx.Error(w, global.ErrArgs.Msg(err.Error()))
			return
		}

		info, err := ctx.AccessSvc.Me(r.Context(), cv)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, info)
	}
}
