package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/handler"
	"github.com/ryanreadbooks/whimer/api-x/internal/middleware"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

// 笔记管理路由
func regNoteRoutes(group *xhttp.RouterGroup, svc *handler.Handler) {
	g := group.Group("/note")
	{
		// 笔记作者相关
		creator := g.Group("/creator", middleware.MustLogin())
		{
			v1g := creator.Group("/v1")
			// 发布笔记
			v1g.Post("/create", svc.Note.CreatorCreateNote())
			// 更新笔记
			v1g.Post("/update", svc.Note.CreatorUpdateNote())
			// 删除笔记
			v1g.Post("/delete", svc.Note.CreatorDeleteNote())
			// 列出笔记
			v1g.Get("/list", svc.Note.CreatorListNotes())
			// 分页列出笔记
			v1g.Get("/list/page", svc.Note.CreatorPageListNotes())
			// 获取笔记
			v1g.Get("/get/:note_id", svc.Note.CreatorGetNote())
			// 申请笔记资源上传链接
			v1g.Get("/upload/auth", svc.Note.CreatorUploadNoteAuth(), middleware.ApiOffline()) // Deprecated
		}
		{
			v2g := creator.Group("/v2")
			v2g.Get("/upload/auth", svc.Note.CreatorUploadNoteAuthV2())
		}

		// 笔记互动相关接口
		interact := g.Group("/interact", middleware.MustLogin())
		{
			v1g := interact.Group("/v1")
			// 点赞/取消点赞笔记
			v1g.Post("/like", svc.Note.LikeNote())
			// 获取笔记点赞数量
			v1g.Get("/likes/:note_id", svc.Note.GetNoteLikeCount())
			// 获取点赞过的笔记
			v1g.Get("/like/notes", svc.Note.GetLikeNotes())
		}
	}
}
