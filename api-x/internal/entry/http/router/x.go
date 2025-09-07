package router

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"

	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler"
	zeroservice "github.com/zeromicro/go-zero/core/service"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterX(engine *rest.Server, h *handler.Handler) {
	root := xhttp.NewRouterGroup(engine)
	xGroup := root.Group("/x")

	// register all routes
	// note routes
	regNoteRoutes(xGroup, h)
	// comment routes
	regCommentRoutes(xGroup, h)
	regProfileRoutes(xGroup, h)
	// relation routes
	regRelationRoutes(xGroup, h)
	regChatRoutes(xGroup, h)
	regUserRoutes(xGroup, h)

	// feed routes
	regFeedRoutes(xGroup, h)
	// search routes
	regSearchRoutes(xGroup, h)

	mod := h.Config.Http.Mode
	if mod == zeroservice.DevMode || mod == zeroservice.TestMode {
		engine.PrintRoutes()
	}
}
