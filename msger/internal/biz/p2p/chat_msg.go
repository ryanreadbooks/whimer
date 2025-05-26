package p2p

// TODO 定义消息的类型
type MsgType int8

// TODO 定义消息的状态
type MsgStatus int8

const (
	MsgStatusNormal  MsgStatus= 0
	MsgStatusRevoked MsgStatus= 1
)

type ChatMsg struct {
	MsgId    int64
	Sender   int64
	Receiver int64
	ChatId   int64
	Type     MsgType
	Status   MsgStatus
	Content  string
	Seq      int64
}

type CreateMsgReq struct {
	ChatId   int64
	Sender   int64
	Receiver int64
	MsgType  MsgType
	Content  string
}
