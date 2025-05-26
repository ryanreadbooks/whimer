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
	LastMessageSeq    int64 `db:"last_message_seq"`
	LastReadMessageId int64 `db:"last_read_message_id"`
	LastReadTime      int64 `db:"last_read_time"`
}

var (
	_chatInst = &Chat{}

	chatFields                = xsql.GetFields(_chatInst)
	insChatFields, insChatQst = xsql.GetFields2(_chatInst, "id") // for insert
)
