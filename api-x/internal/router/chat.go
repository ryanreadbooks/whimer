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
			// 发起会话
			v1g.Post("/chat/create", svc.Chat.CreateChat())
			// 拉取消息列表
			v1g.Get("/chat/list", svc.Chat.ListChats())
			// 获取会话
			v1g.Get("/chat", svc.Chat.GetChat())
			// 拉消息
			v1g.Get("/message/list", svc.Chat.ListMessages())
		}
	}
}
