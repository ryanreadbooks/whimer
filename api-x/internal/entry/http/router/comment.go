package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/middleware"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

// 评论路由
func regCommentRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	g := group.Group("/comment", middleware.MustLogin())
	{
		v1NoteGroup := g.Group("/v1/note")
		{
			// 发布评论
			v1NoteGroup.Post("/pub", h.Comment.PublishComment())
			// 获取主评论
			v1NoteGroup.Get("/roots", h.Comment.PageGetRoots())
			// 获取子评论
			v1NoteGroup.Get("/subs", h.Comment.PageGetSubs())
			// 获取主评论
			v1NoteGroup.Get("/replies", h.Comment.PageGetReplies())
			// 删除评论
			v1NoteGroup.Post("/del", h.Comment.DelComment())
			// 置顶评论
			v1NoteGroup.Post("/pin", h.Comment.PinComment())
			// 点赞评论
			v1NoteGroup.Post("/like", h.Comment.LikeComment())
			// 点踩评论
			v1NoteGroup.Post("/dislike", h.Comment.DislikeComment())
			// 获取评论点赞数
			v1NoteGroup.Get("/likes", h.Comment.GetCommentLikeCount())
		}
	}
}
