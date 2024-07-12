package handler

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/note/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// 注册路由
func RegisterHandlers(engine *rest.Server, ctx *svc.ServiceContext) {
	engine.AddRoutes(routes(ctx), rest.WithPrefix("/note"))
}

func routes(ctx *svc.ServiceContext) []rest.Route {
	rs := make([]rest.Route, 0)
	rs = append(rs, noteCreatorRoutes(ctx)...)

	return rs
}

// 笔记管理路由
func noteCreatorRoutes(ctx *svc.ServiceContext) []rest.Route {
	return []rest.Route{
		xhttp.Post("/v1/creator/create", CreatorCreateHandler(ctx)),
		xhttp.Post("/v1/creator/update", CreatorUpdateHandler(ctx)),
		xhttp.Post("/v1/creator/delete", CreatorDeleteHandler(ctx)),
		xhttp.Get("/v1/creator/list", CreatorListHandler(ctx)),
		xhttp.Get("/v1/creator/get/:note_id", CreatorGetNoteHandler(ctx)),
		xhttp.Get("/v1/creator/upload/auth", UploadAuthHandler(ctx)),
	}
}
