package p2p

import (
	"github.com/ryanreadbooks/whimer/misc/xsql"
	gm "github.com/ryanreadbooks/whimer/msger/internal/global/model"
)

type MessagePO struct {
	Id       int64        `db:"id"`
	MsgId    int64        `db:"msg_id"`
	SenderId int64        `db:"sender_id"`
	ChatId   int64        `db:"chat_id"`
	MsgType  gm.MsgType   `db:"msg_type"`
	Content  string       `db:"content"`
	Status   gm.MsgStatus `db:"status"`
	Seq      int64        `db:"seq"`
	Utime    int64        `db:"utime"`
}

var (
	_msgInst = &MessagePO{}

	msgFields               = xsql.GetFields(_msgInst)
	insMsgFields, insMsgQst = xsql.GetFields2(_msgInst, "id") // for insert
)
