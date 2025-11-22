package chat

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

const (
	msgPOTableName = "message"
)

var (
	msgPOFields = xsql.GetFieldSlice(&MsgPO{})
)

// message
type MsgPO struct {
	Id      uuid.UUID       `db:"id"`
	Type    model.MsgType   `db:"type"`
	Status  model.MsgStatus `db:"status"`
	Sender  int64           `db:"sender"`
	Mtime   int64           `db:"mtime"`
	Content []byte          `db:"content"`
	Ext     int8            `db:"ext"` // 0-noExt; 1-hasExt
	Cid     string          `db:"cid"` // 客户端侧消息id
}

func (m *MsgPO) Values() []any {
	return []any{
		m.Id,
		m.Type,
		m.Status,
		m.Sender,
		m.Mtime,
		m.Content,
		m.Ext,
		m.Cid,
	}
}

func (MsgPO) TableName() string {
	return msgPOTableName
}
