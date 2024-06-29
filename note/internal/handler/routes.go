package handler

import (
	"net/http"

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
	rs = append(rs, noteUploadAuthRoutes(ctx)...)

	return rs
}

// 笔记管理路由
func noteManageRoutes(ctx *svc.ServiceContext) []rest.Route {
	return []rest.Route{
		{
			Method:  http.MethodPost,
			Path:    "/v1/manage/create",
			Handler: ManageCreateHandler(ctx),
		},
		{
			Method:  http.MethodPost,
			Path:    "/v1/manage/update",
			Handler: ManageUpdateHandler(ctx),
		},
	}
}

// 笔记资源上传凭证获取路由
func noteUploadAuthRoutes(ctx *svc.ServiceContext) []rest.Route {
	return []rest.Route{
		{
			Method:  http.MethodGet,
			Path:    "/v1/upload/auth",
			Handler: UploadAuthHandler(ctx),
		},
	}
}
