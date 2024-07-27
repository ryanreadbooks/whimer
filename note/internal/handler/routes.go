package handler

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xhttp/middleware/auth"
	"github.com/ryanreadbooks/whimer/note/internal/external"
	"github.com/ryanreadbooks/whimer/note/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

// 注册路由
func RegisterHandlers(engine *rest.Server, ctx *svc.ServiceContext) {
	xGroup := xhttp.NewRouterGroup(engine)
	regCreatorRoutes(xGroup, ctx)

	mod := ctx.Config.Http.Mode
	if mod == service.DevMode || mod == service.TestMode {
		engine.PrintRoutes()
	}
}

// 笔记管理路由
func regCreatorRoutes(group *xhttp.RouterGroup, ctx *svc.ServiceContext) {
	creatorGroup := group.Group("/creator", auth.UserWeb(external.GetAuther()))
	{
		v1Group := creatorGroup.Group("/v1/note")
		{
			v1Group.Post("/create", NoteCreateHandler(ctx))
			v1Group.Post("/update", NoteUpdateHandler(ctx))
			v1Group.Post("/delete", NoteDeleteHandler(ctx))
			v1Group.Get("/list", NotListHandler(ctx))
			v1Group.Get("/get/:note_id", NoteGetNoteHandler(ctx))
			v1Group.Get("/upload/auth", UploadAuthHandler(ctx))
		}

	}
}
