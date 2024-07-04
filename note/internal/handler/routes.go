package handler

import (
	uhttp "github.com/ryanreadbooks/whimer/misc/utils/http"
	"github.com/ryanreadbooks/whimer/note/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

// 注册路由
func RegisterHandlers(engine *rest.Server, ctx *svc.ServiceContext) {
	engine.AddRoutes(routes(ctx), rest.WithPrefix("/note"))
}

func routes(ctx *svc.ServiceContext) []rest.Route {
	rs := make([]rest.Route, 0)
	rs = append(rs, noteManageRoutes(ctx)...)

	return rs
}

// 笔记管理路由
func noteManageRoutes(ctx *svc.ServiceContext) []rest.Route {
	return []rest.Route{
		uhttp.Post("/v1/manage/create", ManageCreateHandler(ctx)),
		uhttp.Post("/v1/manage/update", ManageUpdateHandler(ctx)),
		uhttp.Post("/v1/manage/delete", ManageDeleteHandler(ctx)),
		uhttp.Get("/v1/manage/list", ManageListHandler(ctx)),
		uhttp.Get("/v1/manage/get/:note_id", ManageGetNoteHandler(ctx)),
		uhttp.Get("/v1/manage/upload/auth", UploadAuthHandler(ctx)),
	}
}
