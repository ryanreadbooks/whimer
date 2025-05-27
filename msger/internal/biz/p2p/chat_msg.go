package p2p

// 定义消息的类型
type MsgType int8

const (
	MsgText  MsgType = 1
	MsgImage MsgType = 10
	MsgVideo MsgType = 20
)

// 定义消息的状态
type MsgStatus int8

const (
	MsgStatusNormal  MsgStatus = 0
	MsgStatusRevoked MsgStatus = 1
)

// 收件箱状态
type InboxStatus int8

const (
	InboxUnread  InboxStatus = 0
	InboxRead    InboxStatus = 1
	InboxRevoked InboxStatus = 2
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
