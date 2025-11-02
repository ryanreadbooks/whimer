package router

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/handler"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/middleware"
)

// 消息路由
func regChatRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	whisperGroup := group.Group("/whisper", middleware.MustLogin())
	{
		v1Group := whisperGroup.Group("/v1")
		{
			// TODO 
			_ = v1Group 
		}
	}

	// 系统消息
	sysMsgGroup := group.Group("/sysmsg", middleware.MustLogin())
	{
		v1Group := sysMsgGroup.Group("/v1")
		{
			v1Group.Post("/chat/read", h.Chat.ClearChatUnread())
			v1Group.Get("/mentions", h.Chat.ListSysMsgMentions())
			v1Group.Get("/replies", h.Chat.ListSysMsgReplies())
		}
	}

	// 聚合的未读数拉取
	msgGroup := group.Group("/msg", middleware.MustLogin())
	{
		v1Group := msgGroup.Group("/v1")
		{
			v1Group.Get("/unread_count", h.Chat.GetTotalUnreadCount())
		}
	}
}
