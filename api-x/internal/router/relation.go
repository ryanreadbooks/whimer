package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/backend"
	"github.com/ryanreadbooks/whimer/api-x/internal/middleware"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

func regRelationRoutes(group *xhttp.RouterGroup, svc *backend.Handler) {
	g := group.Group("/relation")
	{
		v1g := g.Group("/v1", middleware.MustLogin())
		{
			// 关注/取关某个用户
			v1g.Post("/follow", svc.Relation.UserFollowAction())
		}
	}
}
