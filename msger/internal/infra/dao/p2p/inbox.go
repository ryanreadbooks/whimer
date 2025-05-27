package p2p

import "github.com/ryanreadbooks/whimer/misc/xsql"

// 用户收件箱
type InboxMsg struct {
	Id     int64 `db:"id"`
	UserId int64 `db:"user_id"`
	ChatId int64 `db:"chat_id"`
	MsgId  int64 `db:"msg_id"`
	Status int8  `db:"status"`
	Ctime  int64 `db:"ctime"`
}

var (
	_inboxInst = &InboxMsg{}

	inboxFields                 = xsql.GetFields(_inboxInst)
	insInboxFields, insInboxQst = xsql.GetFields2(_inboxInst, "id") // for insert
)
