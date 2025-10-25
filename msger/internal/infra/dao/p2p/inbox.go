package p2p

import (
	"github.com/ryanreadbooks/whimer/misc/xsql"
	gm "github.com/ryanreadbooks/whimer/msger/internal/model"
)

// 用户收件箱
type InboxMsgPO struct {
	Id     int64             `db:"id"`
	UserId int64             `db:"user_id"`
	ChatId int64             `db:"chat_id"`
	MsgId  int64             `db:"msg_id"`
	MsgSeq int64             `db:"msg_seq"`
	Status gm.P2PInboxStatus `db:"status"`
	Ctime  int64             `db:"ctime"`
}

var (
	_inboxInst = &InboxMsgPO{}

	inboxFields                 = xsql.GetFields(_inboxInst)
	insInboxFields, insInboxQst = xsql.GetFields2WithSkip(_inboxInst, "id") // for insert
)
