package http

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xhttp/middleware/csrf"
	"github.com/ryanreadbooks/whimer/passport/internal/config"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
	"github.com/ryanreadbooks/whimer/passport/internal/srv"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取手机验证码
func SmsSendHandler(serv *srv.Service) http.HandlerFunc {
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

		err = serv.AccessSrv.SendSmsCode(r.Context(), req.Tel)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

// 手机号+短信验证码登录
func CheckInWithSmsHandler(serv *srv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[model.SmsCheckInRequest](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		data, err := serv.AccessSrv.SmsCheckIn(r.Context(), req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		sessCookie := data.Session.Cookie()
		http.SetCookie(w, sessCookie)
		http.SetCookie(w, data.Session.UidCookie())
		http.SetCookie(w, csrf.GetToken().Cookie(config.Conf.Domain, sessCookie.Expires))
		w.Header().Add("Vary", "Cookie")
		httpx.OkJson(w, data)
	}
}

// 针对当前session退出登录
func CheckOutCurrentHandler(serv *srv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO do some checking
		sessId := model.CtxGetSessId(r.Context())
		err := serv.AccessSrv.CheckOutCurrent(r.Context(), sessId)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		// invalidate cookie
		http.SetCookie(w, expiredSessIdCookie())
		http.SetCookie(w, csrf.Invalidate())
		httpx.Ok(w)
	}
}

// 全平台退登
func CheckOutAllPlatformHandler(serv *srv.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := serv.AccessSrv.CheckoutAll(r.Context())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		http.SetCookie(w, expiredSessIdCookie())
		http.SetCookie(w, csrf.Invalidate())
		httpx.Ok(w)
	}
}

func expiredSessIdCookie() *http.Cookie {
	return &http.Cookie{
		Name:   model.WhimerSessId,
		Value:  "",
		MaxAge: -1,
	}
}
