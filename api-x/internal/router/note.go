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
			// 发布笔记
			v1g.Post("/create", svc.CreateNote())
			// 更新笔记
			v1g.Post("/update", svc.UpdateNote())
			// 删除笔记
			v1g.Post("/delete", svc.DeleteNote())
			// 列出笔记
			v1g.Get("/list", svc.ListNotes())
			// 获取笔记
			v1g.Get("/get/:note_id", svc.GetNote())
			// 申请笔记资源上传链接
			v1g.Get("/upload/auth", svc.UploadNoteAuth())
			// 点赞/取消点赞笔记
			v1g.Post("/like", svc.LikeNote())
			// 获取笔记点赞数量
			v1g.Get("/likes/:note_id", svc.GetNoteLikeCount())
		}
	}
}
