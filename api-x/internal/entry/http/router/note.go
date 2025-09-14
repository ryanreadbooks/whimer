package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler"
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/middleware"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
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
				// 申请笔记资源上传链接
				v1g.Get("/upload/auth", h.Note.CreatorUploadNoteAuth(), middleware.ApiOffline()) // Deprecated
			}
		}
		{
			v2g := noteCreatorGroup.Group("/v2")
			{
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
	}
}
