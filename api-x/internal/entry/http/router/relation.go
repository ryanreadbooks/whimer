package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/middleware"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

func regRelationRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	g := group.Group("/relation")
	{
		v1g := g.Group("/v1", middleware.MustLogin())
		{
			// 关注/取关某个用户
			v1g.Post("/follow", h.Relation.UserFollowAction())
			// 检查是否关注了某个用户
			v1g.Get("/is_following", h.Relation.GetIsFollowing())
		}
	}
}
