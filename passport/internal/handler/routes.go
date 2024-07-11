package handler

import (
	uhttp "github.com/ryanreadbooks/whimer/misc/utils/http"
	"github.com/ryanreadbooks/whimer/passport/internal/handler/middleware"
	"github.com/ryanreadbooks/whimer/passport/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// 注册路由
func RegisterHandlers(engine *rest.Server, ctx *svc.ServiceContext) {
	registerPassportRoutes(engine, ctx)
	registerProfileRoutes(engine, ctx)
}

func registerPassportRoutes(engine *rest.Server, ctx *svc.ServiceContext) {
	engine.AddRoutes(signInRoutes(ctx), rest.WithPrefix("/passport"))
}

func signInRoutes(ctx *svc.ServiceContext) []rest.Route {
	return []rest.Route{
		uhttp.Post("/v1/sms/send", SmsSendHandler(ctx)),  // 获取登录短信验证码
		uhttp.Post("/v1/signin/sms", SignInWithSms(ctx)), // 手机号+短信验证码登录
	}
}

func registerProfileRoutes(engine *rest.Server, ctx *svc.ServiceContext) {
	// 中间件定义
	middlewares := []rest.Middleware{
		middleware.EnsureSignedIn(ctx),
	}
	engine.AddRoutes(
		rest.WithMiddlewares(
			middlewares,
			profileRoutes(ctx)...,
		),
		rest.WithPrefix("/profile"),
	)
}

func profileRoutes(ctx *svc.ServiceContext) []rest.Route {
	return []rest.Route{
		uhttp.Get("/v1/me", ProfileMe(ctx)),               // 获取个人信息
		uhttp.Post("/v1/me/update", ProfileUpdateMe(ctx)), // 更新个人信息
	}
}
