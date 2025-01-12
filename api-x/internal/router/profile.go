package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/backend"
	"github.com/ryanreadbooks/whimer/api-x/internal/middleware"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

// 用户信息相关的一些接口
func regProfileRoutes(group *xhttp.RouterGroup, svc *backend.Handler) {
	g := group.Group("/profile")
	{
		v1g := g.Group("/v1", middleware.MustLogin())
		{
			// 获取用户的投稿数量、点赞数量等信息
			v1g.Get("/stat", svc.GetProfileStat())
		}
	}
}
