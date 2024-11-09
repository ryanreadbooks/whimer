package http

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xhttp"
	global "github.com/ryanreadbooks/whimer/passport/internal/global"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
	"github.com/ryanreadbooks/whimer/passport/internal/srv"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取手机验证码
func SmsSendHandler(ctx *srv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[model.SendSmsRequest](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		if err := req.Validate(); err != nil {
			httpx.Error(w, err)
			return
		}

		err = ctx.AccessSrv.SendSmsCode(r.Context(), req.Tel)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

// 手机号+短信验证码登录
func CheckInWithSmsHandler(ctx *srv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[model.SmsCheckInRequest](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		data, err := ctx.AccessSrv.SmsCheckIn(r.Context(), req)
		if err != nil {
			xhttp.Error(r,w,err)
			return
		}

		http.SetCookie(w, data.Session.Cookie())
		httpx.OkJson(w, data)
	}
}

// 针对当前session退出登录
func CheckOutCurrentHandler(ctx *srv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO do some checking
		sessId := model.CtxGetSessId(r.Context())
		err := ctx.AccessSrv.CheckOutCurrent(r.Context(), sessId)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		// invalidate cookie
		http.SetCookie(w, expiredSessIdCookie())
		httpx.Ok(w)
	}
}

// 全平台退登
func CheckOutAllPlatformHandler(ctx *srv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		httpx.Error(w, global.ErrApiUnimplemented)
	}
}

func expiredSessIdCookie() *http.Cookie {
	return &http.Cookie{
		Name:   model.WhimerSessId,
		Value:  "",
		MaxAge: -1,
	}
}
