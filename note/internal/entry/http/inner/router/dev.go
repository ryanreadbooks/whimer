package router

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/note/internal/entry/http/inner/handler"

	"github.com/zeromicro/go-zero/rest"
)

func rootGroup(engine *rest.Server) *xhttp.RouterGroup {
	root := xhttp.NewRouterGroup(engine)
	rootGroup := root.Group("/inner")

	return rootGroup
}

// RegisterInner 注册内部接口路由
func RegisterInner(engine *rest.Server, h *handler.Handler) {
	rg := rootGroup(engine)
	apiV1Dev := rg.Group("/api/v1/dev")

	// register all routes
	regNoteProcessRoutes(apiV1Dev, h)
}
