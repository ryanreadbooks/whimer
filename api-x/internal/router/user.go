package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/handler"
	"github.com/ryanreadbooks/whimer/api-x/internal/middleware"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

// 用户信息相关路由由
func regUserRoutes(group *xhttp.RouterGroup, svc *handler.Handler) {
	g := group.Group("/user", middleware.MustLogin())
	{
		v1g := g.Group("/v1")
		{
			// 拉取消息列表
			v1g.Get("/info/list", svc.User.ListInfos())
		}
	}
}
