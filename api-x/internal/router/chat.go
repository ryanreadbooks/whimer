package router

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/backend"
	"github.com/ryanreadbooks/whimer/api-x/internal/middleware"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
)

// 消息路由
func regChatRoutes(group *xhttp.RouterGroup, svc *backend.Handler) {
	g := group.Group("/msg", middleware.MustLogin())
	{
		v1g := g.Group("/v1")
		{
			// 拉取消息列表
			v1g.Get("/chat/list", svc.Chat.ListChats())
			// 拉消息
			v1g.Get("/message/list", svc.Chat.ListMessages())
		}
	}
}
