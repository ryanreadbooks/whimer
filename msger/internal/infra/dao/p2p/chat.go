package p2p

type ChatPO struct {
	Id            int64 `db:"id"`
	ChatId        int64 `db:"chat_id"`
	UserId        int64 `db:"user_id"`
	PeerId        int64 `db:"peer_id"`
	UnreadCount   int64 `db:"unread_count"`
	Ctime         int64 `db:"ctime"`
	LastMsgId     int64 `db:"last_msg_id"` // last_message可以是对方的也可以是自己的
	LastMsgSeq    int64 `db:"last_msg_seq"`
	LastReadMsgId int64 `db:"last_read_msg_id"` // 最后已读的消息id
	LastReadTime  int64 `db:"last_read_time"`   // 消除未读数的时间
}
