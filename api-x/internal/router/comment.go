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
			// 发布评论
			v1g.Post("/pub", svc.PublishComment())
			v1g.Get("/roots", svc.PageGetRoots())
			// 获取子评论
			v1g.Get("/subs", svc.PageGetSubs())
			// 获取主评论
			v1g.Get("/replies", svc.PageGetReplies())
			// 删除评论
			v1g.Post("/del", svc.DelComment())
			// 置顶评论
			v1g.Post("/pin", svc.PinComment())
		}
	}
}
