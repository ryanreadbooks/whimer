package chat

import (
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

const (
	msgExtPOTableName = "message_ext"
)

var (
	msgExtPOFields = xsql.GetFieldSlice(&MsgExtPO{})
)

type MsgExtPO struct {
	MsgId     uuid.UUID       `db:"msg_id"`
	ImageKeys json.RawMessage `db:"image_keys"`
}

func (MsgExtPO) TableName() string {
	return msgExtPOTableName
}

func (m *MsgExtPO) Values() []any {
	return []any{
		m.MsgId,
		m.ImageKeys,
	}
}
