package handler

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/note/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(engine *rest.Server, ctx *svc.ServiceContext) {
	engine.AddRoutes(routes(ctx), rest.WithPrefix("/note"))
}

func routes(ctx *svc.ServiceContext) []rest.Route {
	rs := make([]rest.Route, 0)
	rs = append(rs, noteManageRoutes(ctx)...)

	return rs
}

func noteManageRoutes(ctx *svc.ServiceContext) []rest.Route {
	return []rest.Route{
		{
			Method:  http.MethodPost,
			Path:    "/v1/manage/create",
			Handler: ManageCreateHandler(ctx),
		},
		{
			Method: http.MethodPost,
			Path: "/v1/manage/update",
			Handler: ManageUpdateHandler(ctx),
		},
	}
}
