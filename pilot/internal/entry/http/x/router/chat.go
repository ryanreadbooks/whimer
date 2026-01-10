package router

import (
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/handler"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/middleware"
)

// 消息路由
func regChatRoutes(group *xhttp.RouterGroup, h *handler.Handler) {
	whisperGroup := group.Group("/whisper", middleware.MustLogin())
	{
		v1Group := whisperGroup.Group("/v1")
		{
			// 发起会话
			v1Group.Post("/chat/create", h.Chat.CreateWhisperChat())
			// 发消息
			v1Group.Post("/chat/msg/create", h.Chat.SendWhisperChatMsg())
			// 撤回消息
			v1Group.Post("/chat/msg/recall", h.Chat.RecallWhisperChatMsg())
			// 获取最近会话列表
			v1Group.Get("/recent/chats", h.Chat.ListWhisperRecentChats())
			// 获取会话消息列表
			v1Group.Get("/chat/msgs", h.Chat.ListWhisperChatMsgs())
			// 清除会话未读数
			v1Group.Post("/chat/unread/clear", h.Chat.ClearWhisperChatUnread())
		}
	}

	// 系统消息
	sysMsgGroup := group.Group("/sysmsg", middleware.MustLogin())
	{
		v1Group := sysMsgGroup.Group("/v1")
		{
			v1Group.Post("/chat/read", h.Chat.ClearSysChatUnread())
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
