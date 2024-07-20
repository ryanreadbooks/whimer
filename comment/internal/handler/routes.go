package handler

import (
	"github.com/ryanreadbooks/whimer/comment/internal/svc"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(engine *rest.Server, ctx *svc.ServiceContext) {
	xGroup := xhttp.NewRouterGroup(engine)
	_ = xGroup

	mod := ctx.Config.Http.Mode
	if mod == service.DevMode || mod == service.TestMode {
		engine.PrintRoutes()
	}
}
