package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/middleware"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

// 用户信息相关路由
func regUserRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	g := group.Group("/user", middleware.MustLogin())
	{
		v1g := g.Group("/v1")
		{
			// 批量拉取用户信息
			v1g.Get("/info/list", h.User.ListInfos())
			// 拉取单个用户的信息
			v1g.Get("/get", h.User.GetUser())
		}
	}
}
