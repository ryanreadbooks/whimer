package p2p

type MsgType int8

type MsgStatus int8

type Message struct {
	Id         int64     `db:"id"`
	MsgId      int64     `db:"msg_id"`
	SenderId   int64     `db:"sender_id"`
	ReceiverId int64     `db:"receiver_id"`
	ChatId     int64     `db:"chat_id"`
	MsgType    MsgType   `db:"msg_type"`
	Content    string    `db:"content"`
	Status     MsgStatus `db:"status"`
	Ctime      int64     `db:"ctime"`
	Utime      int64     `db:"utime"`
}
