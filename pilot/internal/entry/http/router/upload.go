package router

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/handler"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/middleware"
)

func regUploadRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	uploadGroup := group.Group("/upload",middleware.MustLogin())
	{
		uploadGroup.Get("/v1/creds", h.Upload.GetTempCreds())
	}
}
