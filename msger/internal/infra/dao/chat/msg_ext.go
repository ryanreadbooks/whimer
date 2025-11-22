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
	MsgId  uuid.UUID       `db:"msg_id"`
	Recall json.RawMessage `db:"recall"` // 消息撤回相关记录
}

func (MsgExtPO) TableName() string {
	return msgExtPOTableName
}

func (m *MsgExtPO) Values() []any {
	// 不能插入nil 当json.RawMessage为nil时赋一个空值
	recall := m.Recall
	if recall == nil {
		recall = json.RawMessage{}
	}

	return []any{
		m.MsgId,
		recall,
	}
}
