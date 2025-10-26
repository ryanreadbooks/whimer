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
			chatGroup := v1Group.Group("/chat")
			{
				chatGroup.Get("", h.Chat.GetChat())            // 获取会话
				chatGroup.Post("/create", h.Chat.CreateChat()) // 发起会话
				chatGroup.Post("/delete", h.Chat.DeleteChat()) // 删除会话
				chatGroup.Get("/list", h.Chat.ListChats())     // 拉取消息列表
			}

			msgGroup := v1Group.Group("/message")
			{
				msgGroup.Get("/list", h.Chat.ListMsgs())     // 拉消息
				msgGroup.Post("/send", h.Chat.SendMsg())     // 发消息
				msgGroup.Post("/revoke", h.Chat.RevokeMsg()) // 撤回消息
				msgGroup.Post("/delete", h.Chat.DeleteMsg()) // 删除消息
			}
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
			v1Group.Get("/likes", h.Chat.ListSysMsgLikes())
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
