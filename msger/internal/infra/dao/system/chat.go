package system

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

type ChatPO struct {
	Id            uuid.UUID            `db:"id"` // uuidv7
	Type          model.SystemChatType `db:"type"`
	Uid           int64                `db:"uid"` // 会话所属用户uid
	Mtime         int64                `db:"mtime"`
	LastMsgId     uuid.UUID            `db:"last_msg_id"`
	LastReadMsgId uuid.UUID            `db:"last_read_msg_id"`
	UnreadCount   int64                `db:"unread_count"`
}

var (
	_systemChatInst = &ChatPO{}

	systemChatFields                      = xsql.GetFields(_systemChatInst)
	insSystemChatFields, insSystemChatQst = xsql.GetFields2WithSkip(_systemChatInst) // for insert
)
