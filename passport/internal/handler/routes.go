package handler

import (
	uhttp "github.com/ryanreadbooks/whimer/misc/utils/http"
	"github.com/ryanreadbooks/whimer/passport/internal/svc"
	"github.com/zeromicro/go-zero/rest"
)

// 注册路由
func RegisterHandlers(engine *rest.Server, ctx *svc.ServiceContext) {
	passportRoutes(engine, ctx)
}

func passportRoutes(engine *rest.Server, ctx *svc.ServiceContext) {
	engine.AddRoutes(signInRoutes(ctx), rest.WithPrefix("/passport"))
}

func signInRoutes(ctx *svc.ServiceContext) []rest.Route {
	return []rest.Route{
		uhttp.Post("/v1/sms/send", SmsSendHandler(ctx)),  // 获取登录短信验证码
		uhttp.Post("/v1/signin/sms", SignInWithSms(ctx)), // 手机号+短信验证码登录
		uhttp.Get("/v1/me", PassportMe(ctx)),             // 获取个人信息
	}
}
