package router

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xhttp/middleware/remote"

	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/handler"
	
	"github.com/zeromicro/go-zero/rest"
)

func rootGroup(engine *rest.Server) *xhttp.RouterGroup {
	root := xhttp.NewRouterGroup(engine)
	rootGroup := root.Group("", remote.ClientAddr)

	return rootGroup
}

// 用户接口
func RegisterX(engine *rest.Server, h *handler.Handler) {
	rg := rootGroup(engine)
	xGroup := rg.Group("/x")

	// register all routes
	// note routes
	regNoteRoutes(xGroup, h)
	// relation routes
	regRelationRoutes(xGroup, h)
	regChatRoutes(xGroup, h)
	regUserRoutes(xGroup, h)

	// feed routes
	regFeedRoutes(xGroup, h)
	// search routes
	regSearchRoutes(xGroup, h)
	// upload routes
	regUploadRoutes(xGroup, h)
}
