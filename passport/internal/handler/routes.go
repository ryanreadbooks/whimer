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
		uhttp.Post("/v1/sms/send", SmsSendHandler(ctx)),
	}
}
