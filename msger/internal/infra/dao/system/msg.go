package system

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/global/model"
)

type MsgPO struct {
	Id           uuid.UUID             `db:"id"` // uuidv7
	SystemChatId uuid.UUID             `db:"system_chat_id"`
	Uid          int64                 `db:"uid"`      // 触发系统消息的uid
	RecvUid      int64                 `db:"recv_uid"` // 接收系统消息的uid
	Status       model.SystemMsgStatus `db:"status"`
	MsgType      model.MsgType         `db:"msg_type"`
	Content      string                `db:"content"`
	Mtime        int64                 `db:"mtime"`
}

var (
	_systemMsgInst = &MsgPO{}

	systemMsgFields                     = xsql.GetFields(_systemMsgInst)
	insSystemMsgFields, insSystemMsgQst = xsql.GetFields2WithSkip(_systemMsgInst) // for insert
)
