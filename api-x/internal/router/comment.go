package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/handler"
	"github.com/ryanreadbooks/whimer/api-x/internal/middleware"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

// 评论路由
func regCommentRoutes(group *xhttp.RouterGroup, svc *handler.Handler) {
	g := group.Group("/comment", middleware.MustLogin())
	{
		v1g := g.Group("/v1")
		{
			// 发布评论
			v1g.Post("/pub", svc.Comment.PublishComment())
			// 获取主评论
			v1g.Get("/roots", svc.Comment.PageGetRoots())
			// 获取子评论
			v1g.Get("/subs", svc.Comment.PageGetSubs())
			// 获取主评论
			v1g.Get("/replies", svc.Comment.PageGetReplies())
			// 删除评论
			v1g.Post("/del", svc.Comment.DelComment())
			// 置顶评论
			v1g.Post("/pin", svc.Comment.PinComment())
			// 点赞评论
			v1g.Post("/like", svc.Comment.LikeComment())
			// 点踩评论
			v1g.Post("/dislike", svc.Comment.DislikeComment())
			// 获取评论点赞数
			v1g.Get("/likes", svc.Comment.GetCommentLikeCount())
		}
	}
}
