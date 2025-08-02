package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/handler"
	"github.com/ryanreadbooks/whimer/api-x/internal/middleware"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

func regRelationRoutes(group *xhttp.RouterGroup, svc *handler.Handler) {
	g := group.Group("/relation")
	{
		v1g := g.Group("/v1", middleware.MustLogin())
		{
			// 关注/取关某个用户
			v1g.Post("/follow", svc.Relation.UserFollowAction())
			// 检查是否关注了某个用户
			v1g.Get("/is_following", svc.Relation.GetIsFollowing())
		}
	}
}
