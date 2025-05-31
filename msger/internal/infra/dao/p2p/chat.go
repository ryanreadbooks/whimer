package p2p

import "github.com/ryanreadbooks/whimer/misc/xsql"

type ChatPO struct {
	Id             int64 `db:"id"`
	ChatId         int64 `db:"chat_id"`
	UserId         int64 `db:"user_id"`
	PeerId         int64 `db:"peer_id"`
	UnreadCount    int64 `db:"unread_count"`
	Ctime          int64 `db:"ctime"`
	LastMessageId  int64 `db:"last_message_id"` // last_message可以是对方的也可以是自己的
	LastMessageSeq int64 `db:"last_message_seq"`
	LastReadTime   int64 `db:"last_read_time"` // 消除未读数的时间
}

var (
	_chatInst = &ChatPO{}

	chatFields                = xsql.GetFields(_chatInst)
	insChatFields, insChatQst = xsql.GetFields2(_chatInst, "id") // for insert
)
