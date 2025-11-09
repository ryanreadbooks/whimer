package userchat

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	chatdao "github.com/ryanreadbooks/whimer/msger/internal/infra/dao/chat"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

type ChatInbox struct {
	Uid           int64
	ChatId        uuid.UUID
	LastMsgId     uuid.UUID
	LastReadMsgId uuid.UUID
	LastReadTime  int64
	UnreadCount   int64
	Ctime         int64
	Mtime         int64
	Status        model.ChatInboxStatus
	IsPinned      bool
}

func makeChatInboxFromPO(p *chatdao.ChatInboxPO) *ChatInbox {
	return &ChatInbox{
		Uid:           p.Uid,
		ChatId:        p.ChatId,
		LastMsgId:     p.LastMsgId,
		LastReadMsgId: p.LastReadMsgId,
		LastReadTime:  p.LastReadTime,
		UnreadCount:   p.UnreadCount,
		Ctime:         p.Ctime,
		Mtime:         p.Mtime,
		Status:        p.Status,
		IsPinned:      p.IsPinned == model.ChatInboxPinned,
	}
}
