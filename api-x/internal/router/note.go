package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/backend"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

// 笔记管理路由
func regNoteRoutes(group *xhttp.RouterGroup, svc *backend.Handler) {
	g := group.Group("/note")
	{
		v1g := g.Group("/v1")
		{
			admin := v1g.Group("/admin")
			{
				// 发布笔记
				admin.Post("/create", svc.AdminCreateNote())
				// 更新笔记
				admin.Post("/update", svc.AdminUpdateNote())
				// 删除笔记
				admin.Post("/delete", svc.AdminDeleteNote())
				// 列出笔记
				admin.Get("/list", svc.AdminListNotes())
				// 获取笔记
				admin.Get("/get/:note_id", svc.AdminGetNote())
				// 申请笔记资源上传链接
				admin.Get("/upload/auth", svc.AdminUploadNoteAuth())
			}

			// 点赞/取消点赞笔记
			v1g.Post("/like", svc.LikeNote())
			// 获取笔记点赞数量
			v1g.Get("/likes/:note_id", svc.GetNoteLikeCount())
			// 获取笔记信息
			v1g.Get("get/:note_id", svc.GetNote())
		}
	}
}
