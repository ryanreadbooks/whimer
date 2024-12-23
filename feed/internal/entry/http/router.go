package http

import (
	"github.com/ryanreadbooks/whimer/feed/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xhttp/middleware"
	zeroservice "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

func Init(engine *rest.Server) {
	root := xhttp.NewRouterGroup(engine)
	root.Use(middleware.Recovery)
	feed := root.Group("/feed")
	initPrivateGroup(feed)
	initPublicGroup(feed)
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
	public := root.Group("", CanLogin())
	apiv1 := public.Group("/v1")
	{
		apiv1.Get("/recommend", feedRecommend())
		apiv1.Get("/detail", feedDetail())
	}
}
