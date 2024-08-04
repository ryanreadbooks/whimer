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
			v1g.Post("/create", svc.CreateNote())
			v1g.Post("/update", svc.UpdateNote())
			v1g.Post("/delete", svc.DeleteNote())
			v1g.Get("/list", svc.ListNotes())
			v1g.Get("/get/:note_id", svc.GetNote())
			v1g.Get("/upload/auth", svc.UploadNoteAuth())
		}
	}
}
