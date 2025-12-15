package router

import (
	"fmt"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/inner/handler"
)

func regDevRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	webhookGroup := group.Group("/api/v1/webhook")
	webhookGroup.Post("/minio", func(w http.ResponseWriter, r *http.Request) {
		// TODO
		fmt.Println(r)
	})
}
