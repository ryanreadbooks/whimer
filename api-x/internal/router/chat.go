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
		v1Group := g.Group("/v1")
		{
			chatGroup := v1Group.Group("/chat")
			{
				// 获取会话
				chatGroup.Get("", svc.Chat.GetChat())
				// 发起会话
				chatGroup.Post("/create", svc.Chat.CreateChat())
				// 删除会话
				chatGroup.Post("/delete", svc.Chat.DeleteChat())
				// 拉取消息列表
				chatGroup.Get("/list", svc.Chat.ListChats())
			}

			msgGroup := v1Group.Group("/message")
			{
				// 拉消息
				msgGroup.Get("/list", svc.Chat.ListMessages())
				// 发消息
				msgGroup.Post("/send", svc.Chat.SendMessage())
				// 删除消息
				msgGroup.Post("/delete", svc.Chat.DeleteMessage())
			}
		}
	}
}
