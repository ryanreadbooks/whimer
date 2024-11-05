package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/backend"
	"github.com/ryanreadbooks/whimer/api-x/internal/middleware"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

// 笔记管理路由
func regNoteRoutes(group *xhttp.RouterGroup, svc *backend.Handler) {
	g := group.Group("/note")
	{
		admin := g.Group("/admin", middleware.MustLogin())
		{
			v1g := admin.Group("/v1")
			// 发布笔记
			v1g.Post("/create", svc.AdminCreateNote())
			// 更新笔记
			v1g.Post("/update", svc.AdminUpdateNote())
			// 删除笔记
			v1g.Post("/delete", svc.AdminDeleteNote())
			// 列出笔记
			v1g.Get("/list", svc.AdminListNotes())
			// 获取笔记
			v1g.Get("/get/:note_id", svc.AdminGetNote())
			// 申请笔记资源上传链接
			v1g.Get("/upload/auth", svc.AdminUploadNoteAuth())
		}

		interact := g.Group("/interact", middleware.MustLogin())
		{
			v1g := interact.Group("/v1")
			// 点赞/取消点赞笔记
			v1g.Post("/like", svc.LikeNote())
			// 获取笔记点赞数量
			v1g.Get("/likes/:note_id", svc.GetNoteLikeCount())
		}

		feed := g.Group("/feed", middleware.CanLogin())
		{
			v1g := feed.Group("/v1")
			// 获取笔记信息
			v1g.Get("get/:note_id", svc.GetNote())
		}
	}
}
