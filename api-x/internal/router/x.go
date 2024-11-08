package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/backend"
	"github.com/ryanreadbooks/whimer/misc/xhttp"

	zeroservice "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

func RegX(engine *rest.Server, svc *backend.Handler) {
	root := xhttp.NewRouterGroup(engine)
	xGroup := root.Group("/x")

	// register all routes
	// note routes
	regNoteRoutes(xGroup, svc)
	// comment routes
	regCommentRoutes(xGroup, svc)

	mod := svc.Config.Http.Mode
	if mod == zeroservice.DevMode || mod == zeroservice.TestMode {
		engine.PrintRoutes()
	}
}
