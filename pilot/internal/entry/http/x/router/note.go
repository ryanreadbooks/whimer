package router

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/handler"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/middleware"
)

// 笔记管理路由
func regNoteRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	noteGroup := group.Group("/note")
	{
		// 笔记作者相关
		// /note/creator
		noteCreatorGroup := noteGroup.Group("/creator", middleware.MustLogin())
		{
			v1g := noteCreatorGroup.Group("/v1")
			{
				// 发布笔记
				v1g.Post("/create", h.Note.CreatorCreateNote())
				// 更新笔记
				v1g.Post("/update", h.Note.CreatorUpdateNote())
				// 删除笔记
				v1g.Post("/delete", h.Note.CreatorDeleteNote())
				// 分页列出笔记
				v1g.Get("/list", h.Note.CreatorPageListNotes())
				// 获取笔记
				v1g.Get("/get/:note_id", h.Note.CreatorGetNote())
			}
		}
		{
			v2g := noteCreatorGroup.Group("/v2")
			{
				// Deprecated
				v2g.Get("/upload/auth", h.Note.CreatorUploadNoteAuthV2())
			}
		}

		// /note/tag
		noteTagGroup := noteGroup.Group("/tag", middleware.MustLogin())
		{
			noteTagGroup.Post("/v1/create", h.Note.AddNewTag())
			noteTagGroup.Post("/v1/search", h.Note.SearchTags())
		}

		// 笔记互动相关接口
		// /note/interact
		noteInteract := noteGroup.Group("/interact", middleware.MustLogin())
		{
			v1g := noteInteract.Group("/v1")
			{
				// 点赞/取消点赞笔记
				v1g.Post("/like", h.Note.LikeNote())
				// 获取笔记点赞数量
				v1g.Get("/like/:note_id/count", h.Note.GetNoteLikeCount())
				// 获取点赞过的笔记
				v1g.Get("/like/history", h.Note.ListLikedNotes())
			}
		}

		// 笔记评论相关接口
		// /note/comment
		noteCommentGroup := noteGroup.Group("/comment", middleware.CanLogin())
		{
			v1Group := noteCommentGroup.Group("/v1")
			v1AuthedGroup := v1Group.Group("", middleware.MustLoginCheck())
			{
				// 发布评论
				v1AuthedGroup.Post("/pub", h.Comment.PublishNoteComment())
				// 获取主评论
				v1AuthedGroup.Get("/roots", h.Comment.PageGetNoteRootComments())
				// 获取子评论
				v1AuthedGroup.Get("/subs", h.Comment.PageGetNoteSubComments())
				// 分页获取评论
				v1AuthedGroup.Get("/pages", h.Comment.PageGetNoteComments())
				// 删除评论
				v1AuthedGroup.Post("/del", h.Comment.DelNoteComment())
				// 置顶评论
				v1AuthedGroup.Post("/pin", h.Comment.PinNoteComment())
				// 点赞评论
				v1AuthedGroup.Post("/like", h.Comment.LikeNoteComment())
				// 点踩评论
				v1AuthedGroup.Post("/dislike", h.Comment.DislikeNoteComment())
				// 获取评论点赞数
				v1AuthedGroup.Get("/likes", h.Comment.GetNoteCommentLikeCount())
				// 评论中插入图片申请上传凭证
				v1AuthedGroup.Get("/upload/images", h.Comment.UploadCommentImages())
			}
		}
	}
}
