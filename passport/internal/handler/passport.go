package handler

import (
	"net/http"

	global "github.com/ryanreadbooks/whimer/passport/internal/gloabl"
	"github.com/ryanreadbooks/whimer/passport/internal/svc"
	tp "github.com/ryanreadbooks/whimer/passport/internal/types/passport"

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
