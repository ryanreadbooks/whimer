package handler

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/passport/internal/handler/middleware"
	"github.com/ryanreadbooks/whimer/passport/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

// 注册路由
func RegisterHandlers(engine *rest.Server, ctx *svc.ServiceContext) {
	xGroup := xhttp.NewRouterGroup(engine)
	regPassportRoutes(xGroup, ctx)
	regProfileRoutes(xGroup, ctx)

	mod := ctx.Config.Http.Mode
	if mod == service.DevMode || mod == service.TestMode {
		engine.PrintRoutes()
	}
}

func regPassportRoutes(group *xhttp.RouterGroup, ctx *svc.ServiceContext) {
	passportGroup := group.Group("/passport")
	{
		passportGroup.Post("/v1/sms/send", SmsSendHandler(ctx))  // 获取登录短信验证码
		passportGroup.Post("/v1/signin/sms", SignInWithSms(ctx)) // 手机号+短信验证码登录
	}
}

func regProfileRoutes(group *xhttp.RouterGroup, ctx *svc.ServiceContext) {
	profileGroup := group.Group("/profile", middleware.EnsureSignedIn(ctx))
	{
		profileGroup.Get("/v1/me", ProfileMe(ctx))                // 获取个人信息
		profileGroup.Post("/v1/me/update", ProfileUpdateMe(ctx))  // 更新个人信息
		profileGroup.Post("/v1/avatar", ProfileUpdateAvatar(ctx)) // 上传头像
	}
}
