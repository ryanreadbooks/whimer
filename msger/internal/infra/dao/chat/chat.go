package chat

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

const chatPOTableName = "chat"

var (
	chatPOFields = xsql.GetFieldSlice(&ChatPO{})
)

type ChatPO struct {
	Id        uuid.UUID        `db:"id"`
	Type      model.ChatType   `db:"type"`
	Name      string           `db:"name"`
	Status    model.ChatStatus `db:"status"`
	Creator   int64            `db:"creator"`
	Mtime     int64            `db:"mtime"`
	LastMsgId uuid.UUID        `db:"last_msg_id"`
	Settings  int64            `db:"settings"`
}

func (ChatPO) TableName() string {
	return chatPOTableName
}

func (c *ChatPO) Values() []any {
	return []any{
		c.Id,
		c.Type,
		c.Name,
		c.Status,
		c.Creator,
		c.Mtime,
		c.LastMsgId,
		c.Settings,
	}
}

type ChatIdAndTypePO struct {
	Id   uuid.UUID      `db:"id"`
	Type model.ChatType `db:"type"`
}
