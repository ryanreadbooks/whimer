package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/backend"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

// 笔记管理路由
func regNoteRoutes(group *xhttp.RouterGroup, svc *backend.Handler) {
	creatorGroup := group.Group("/note")
	{
		v1Group := creatorGroup.Group("/v1")
		{
			v1Group.Post("/create", svc.CreateNote())
			v1Group.Post("/update", svc.UpdateNote())
			v1Group.Post("/delete", svc.DeleteNote())
			v1Group.Get("/list", svc.ListNotes())
			v1Group.Get("/get/:note_id", svc.GetNote())
			v1Group.Get("/upload/auth", svc.UploadNoteAuth())
		}
	}
}
