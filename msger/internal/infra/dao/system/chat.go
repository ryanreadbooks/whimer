package system

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/global/model"
)

type SystemChatPO struct {
	Id            uuid.UUID            `db:"id"` // uuidv7
	Type          model.SystemChatType `db:"type"`
	Uid           int64                `db:"uid"`
	Mtime         int64                `db:"mtime"`
	LastMsgId     uuid.UUID            `db:"last_msg_id"`
	LastReadMsgId uuid.UUID            `db:"last_read_msg_id"`
	LastReadTime  int64                `db:"last_read_time"`
	UnreadCount   int64                `db:"unread_count"`
}

var (
	_systemChatInst = &SystemChatPO{}

	systemChatFields                      = xsql.GetFields(_systemChatInst)
	insSystemChatFields, insSystemChatQst = xsql.GetFields2WithSkip(_systemChatInst) // for insert
)
