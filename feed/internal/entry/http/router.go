package http

import (
	"github.com/ryanreadbooks/whimer/feed/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	zeroservice "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

func Init(engine *rest.Server) {
	root := xhttp.NewRouterGroup(engine)

	initPrivateGroup(root)
	initPublicGroup(root)
	mod := config.Conf.Http.Mode
	if mod == zeroservice.DevMode || mod == zeroservice.TestMode {
		engine.PrintRoutes()
	}
}

func initPrivateGroup(root *xhttp.RouterGroup) {
	private := root.Group("", MustLogin())
	api := private.Group("/api")
	_ = api
}

func initPublicGroup(root *xhttp.RouterGroup) {
	public := root.Group("", MustLogin())
	api := public.Group("/api")
	apiv1 := api.Group("/v1")
	{
		feed := apiv1.Group("/feed")
		{
			feed.Get("/recommend", feedRecommend())
			feed.Get("/detail", feedDetail())
		}
	}
}
