package chat

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

const (
	chatMsgPOTableName = "chat_message"
)

var (
	chatMsgPOFields = xsql.GetFieldSlice(&ChatMsgPO{})
)

// 会话-消息关联
//
// 一个会话包含多条消息
type ChatMsgPO struct {
	ChatId uuid.UUID `db:"chat_id"`
	MsgId  uuid.UUID `db:"msg_id"`
	Ctime  int64     `db:"ctime"`
	Pos    int64     `db:"pos"` // 会话中消息位置 在单个会话中唯一且单调递增
}

func (ChatMsgPO) TableName() string {
	return chatMsgPOTableName
}

func (p *ChatMsgPO) Values() []any {
	return []any{
		p.ChatId,
		p.MsgId,
		p.Ctime,
		p.Pos,
	}
}
