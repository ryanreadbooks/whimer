package handler

import (
	"net/http"
	"regexp"

	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/svc"
	tp "github.com/ryanreadbooks/whimer/passport/internal/types/passport"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var (
	telRegx = regexp.MustCompile(`^1[3-9]\d{9}$`)
)

func SmsSendHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req tp.SmsSendReq
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.Error(w, global.ErrArgs.Msg(err.Error()))
			return
		}

		if !telRegx.MatchString(req.Tel) {
			httpx.Error(w, global.ErrInvalidTel)
			return
		}

		err := ctx.SignInUpSvc.RequestSms(r.Context(), req.Tel)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, tp.SmdSendRes{})
	}
}
