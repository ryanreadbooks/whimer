package http

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/passport/internal/config"
	"github.com/ryanreadbooks/whimer/passport/internal/entry/http/middleware"
	"github.com/ryanreadbooks/whimer/passport/internal/srv"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

// 注册路由
func Init(engine *rest.Server, ctx *srv.Service) {
	xGroup := xhttp.NewRouterGroup(engine)
	regPassportRoutes(xGroup, ctx)
	regProfileRoutes(xGroup, ctx)

	mod := config.Conf.Http.Mode
	if mod == service.DevMode || mod == service.TestMode {
		engine.PrintRoutes()
	}
}

func regPassportRoutes(group *xhttp.RouterGroup, ctx *srv.Service) {
	passportGroup := group.Group("/passport")
	{
		passportGroup.Post("/v1/sms/send", SmsSendHandler(ctx))           // 获取登录短信验证码
		passportGroup.Post("/v1/checkin/sms", CheckInWithSmsHandler(ctx)) // 手机号+短信验证码登录

		signoutGroup := passportGroup.Group("/v1/checkout", middleware.EnsureCheckedIn(ctx))
		{
			signoutGroup.Post("/current", CheckOutCurrentHandler(ctx)) // 退登
			signoutGroup.Post("/all", CheckOutAllPlatformHandler(ctx)) // 退登全平台
		}
	}
}

func regProfileRoutes(group *xhttp.RouterGroup, ctx *srv.Service) {
	profileGroup := group.Group("/profile", middleware.EnsureCheckedIn(ctx))
	{
		profileGroup.Get("/v1/me", GetMyProfileHandler(ctx))            // 获取个人信息
		profileGroup.Post("/v1/me/update", UpdateMyProfileHandler(ctx)) // 更新个人信息
		profileGroup.Post("/v1/avatar", UpdateMyAvatarHandler(ctx))     // 上传头像
	}
}
