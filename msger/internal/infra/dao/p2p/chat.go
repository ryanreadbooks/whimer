package p2p

import "github.com/ryanreadbooks/whimer/misc/xsql"

type Chat struct {
	Id                int64 `db:"id"`
	ChatId            int64 `db:"chat_id"`
	UserId            int64 `db:"user_id"`
	PeerId            int64 `db:"peer_id"`
	UnReadCount       int64 `db:"unread_count"`
	Ctime             int64 `db:"ctime"`
	LastMessageId     int64 `db:"last_message_id"`
	LastMessageTime   int64 `db:"last_message_time"`
	LastReadMessageId int64 `db:"last_read_message_id"`
	LastReadTime      int64 `db:"last_read_time"`
}

var (
	_inst = &Chat{}

	dbFields          = xsql.GetFields(_inst)
	insFields, insQst = xsql.GetFields2(_inst, "id") // for insert
)
