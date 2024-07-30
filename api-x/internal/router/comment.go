package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/backend"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

// 评论路由
func regCommentRoutes(group *xhttp.RouterGroup, svc *backend.Handler) {
	g := group.Group("/comment")
	{
		v1g := g.Group("/v1")
		{
			v1g.Post("/pub", svc.PublishComment())
			v1g.Get("/get", svc.PageGetComments())
		}
	}
}
