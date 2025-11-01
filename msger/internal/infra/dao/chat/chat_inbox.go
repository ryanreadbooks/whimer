package chat

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

const (
	chatInboxPOTableName = "chat_inbox"
)

var (
	chatInboxPOFields = xsql.GetFieldSlice(&ChatInboxPO{})
)

// 用户收信箱
//
// 收件箱记录了uid在chatId中的消息摘要记录, 不记录收件箱中每条消息具体情况
type ChatInboxPO struct {
	Uid           int64                 `db:"uid"`
	ChatId        uuid.UUID             `db:"chat_id"`
	LastMsgId     uuid.UUID             `db:"last_msg_id"`      // 最先一条消息
	LastReadMsgId uuid.UUID             `db:"last_read_msg_id"` // 最新一条已读消息
	LastReadTime  int64                 `db:"last_read_time"`   // 最后一条已读消息读取时间
	UnreadCount   int64                 `db:"unread_count"`     // 未读数
	Ctime         int64                 `db:"ctime"`
	Mtime         int64                 `db:"mtime"`
	Status        model.ChatInboxStatus `db:"status"`
	IsPinned      int8                  `db:"is_pinned"` // 是否置顶
}

func (ChatInboxPO) TableName() string {
	return chatInboxPOTableName
}

func (p *ChatInboxPO) Values() []any {
	return []any{
		p.Uid,
		p.ChatId,
		p.LastMsgId,
		p.LastReadMsgId,
		p.LastReadTime,
		p.UnreadCount,
		p.Ctime,
		p.Mtime,
		p.Status,
		p.IsPinned,
	}
}
