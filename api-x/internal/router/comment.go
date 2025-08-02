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
		v1NoteGroup := g.Group("/v1/note")
		{
			// 发布评论
			v1NoteGroup.Post("/pub", svc.Comment.PublishComment())
			// 获取主评论
			v1NoteGroup.Get("/roots", svc.Comment.PageGetRoots())
			// 获取子评论
			v1NoteGroup.Get("/subs", svc.Comment.PageGetSubs())
			// 获取主评论
			v1NoteGroup.Get("/replies", svc.Comment.PageGetReplies())
			// 删除评论
			v1NoteGroup.Post("/del", svc.Comment.DelComment())
			// 置顶评论
			v1NoteGroup.Post("/pin", svc.Comment.PinComment())
			// 点赞评论
			v1NoteGroup.Post("/like", svc.Comment.LikeComment())
			// 点踩评论
			v1NoteGroup.Post("/dislike", svc.Comment.DislikeComment())
			// 获取评论点赞数
			v1NoteGroup.Get("/likes", svc.Comment.GetCommentLikeCount())
		}
	}
}
