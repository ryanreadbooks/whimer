package router

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xhttp/middleware/remote"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/inner/handler"

	"github.com/zeromicro/go-zero/rest"
)

func rootGroup(engine *rest.Server) *xhttp.RouterGroup {
	root := xhttp.NewRouterGroup(engine)
	rootGroup := root.Group("", remote.ClientAddr)

	return rootGroup
}

// 开发者接口
func RegisterInner(engine *rest.Server, h *handler.Handler) {
	rg := rootGroup(engine)
	devGroup := rg.Group("/inner/dev")

	// register all routes
	regDevRoutes(devGroup, h)
}
