package handler

import (
	"net/http"

	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
	tp "github.com/ryanreadbooks/whimer/passport/internal/model/passport"
	"github.com/ryanreadbooks/whimer/passport/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 获取手机验证码
func SmsSendHandler(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req tp.SmsSendReq
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.Error(w, global.ErrArgs.Msg(err.Error()))
			return
		}

		if err := req.Validate(); err != nil {
			httpx.Error(w, err)
			return
		}

		err := ctx.AccessSvc.RequestSms(r.Context(), req.Tel)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, tp.SmdSendRes{})
	}
}

// 手机号+短信验证码登录
func SignInWithSms(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req tp.SignInSmdReq
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.Error(w, global.ErrArgs.Msg(err.Error()))
			return
		}

		if err := req.Validate(); err != nil {
			httpx.Error(w, err)
			return
		}

		user, sess, err := ctx.AccessSvc.SignInWithSms(r.Context(), &req)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		http.SetCookie(w, sess.Cookie())
		httpx.OkJson(w, tp.NewFromRepoBasic(user))
	}
}

// 针对当前session退出登录
func SignoutCurrent(ctx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO do some checking
		sessId := model.CtxGetSessId(r.Context())
		err := ctx.AccessSvc.SignOutCurrent(r.Context(), sessId)
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
func SignoutAllPlatform(ctx *svc.ServiceContext) http.HandlerFunc {
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
