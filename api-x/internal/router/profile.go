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
		v1gLogin := g.Group("/v1", middleware.MustLogin())
		{
			// 获取用户的投稿数量、点赞数量等信息
			v1gLogin.Get("/stat", svc.GetProfileStat())
		}

		v1gNoLogin := g.Group("/v1", middleware.CanLogin())
		{
			v1gNoLogin.Get("/hover", svc.GetHoverProfile())
		}
	}
}
