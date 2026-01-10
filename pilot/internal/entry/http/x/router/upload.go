package router

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/handler"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/middleware"
)

func regUploadRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	uploadGroup := group.Group("/upload", middleware.MustLogin())
	{
		// 获取临时上传ak/sk凭证
		uploadGroup.Get("/v1/creds", h.Upload.GetTemporaryCreds())

		// 获取post policy上传凭证
		uploadGroup.Get("/v1/post_policy/creds", h.Upload.GetPostPolicyCreds())
	}
}
