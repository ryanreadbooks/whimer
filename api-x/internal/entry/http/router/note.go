package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler"
	middleware2 "github.com/ryanreadbooks/whimer/api-x/internal/entry/http/middleware"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

// 笔记管理路由
func regNoteRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	g := group.Group("/note")
	{
		// 笔记作者相关
		creator := g.Group("/creator", middleware2.MustLogin())
		{
			v1g := creator.Group("/v1")
			{
				// 发布笔记
				v1g.Post("/create", h.Note.CreatorCreateNote())
				// 更新笔记
				v1g.Post("/update", h.Note.CreatorUpdateNote())
				// 删除笔记
				v1g.Post("/delete", h.Note.CreatorDeleteNote())
				// 列出笔记
				// v1g.Get("/list", svc.Note.CreatorListNotes())
				// 分页列出笔记
				v1g.Get("/list", h.Note.CreatorPageListNotes())
				// 获取笔记
				v1g.Get("/get/:note_id", h.Note.CreatorGetNote())
				// 申请笔记资源上传链接
				v1g.Get("/upload/auth", h.Note.CreatorUploadNoteAuth(), middleware2.ApiOffline()) // Deprecated
			}
		}
		{
			v2g := creator.Group("/v2")
			{
				v2g.Get("/upload/auth", h.Note.CreatorUploadNoteAuthV2())
			}
		}

		tag := g.Group("/tag", middleware2.MustLogin())
		{
			tag.Post("/v1/create", h.Note.AddNewTag())
			tag.Post("/v1/search", h.Note.SearchTags())
		}

		// 笔记互动相关接口
		interact := g.Group("/interact", middleware2.MustLogin())
		{
			v1g := interact.Group("/v1")
			{
				// 点赞/取消点赞笔记
				v1g.Post("/like", h.Note.LikeNote())
				// 获取笔记点赞数量
				v1g.Get("/likes/:note_id", h.Note.GetNoteLikeCount())
				// 获取点赞过的笔记
				v1g.Get("/like/notes", h.Note.GetLikeNotes())
			}
		}
	}
}
