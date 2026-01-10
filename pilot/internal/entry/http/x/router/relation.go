package router

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/handler"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/middleware"
)

func regRelationRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	g := group.Group("/relation", middleware.CanLogin())
	{
		v1g := g.Group("/v1")
		{
			authed := v1g.Group("", middleware.MustLoginCheck())
			// 关注/取关某个用户
			authed.Post("/follow", h.Relation.UserFollowAction())

			// 检查是否关注了某个用户
			v1g.Get("/is_following", h.Relation.GetIsFollowing())
		}
	}
}
