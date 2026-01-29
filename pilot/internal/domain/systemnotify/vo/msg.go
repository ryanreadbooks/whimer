package vo

type SystemMessage struct {
	Uid       int64  // 发送者
	TargetUid int64  // 接受者
	Content   []byte // 内容
}

// RawSystemMsg 原始系统消息
type RawSystemMsg struct {
	Id      string
	RecvUid int64
	Content []byte
	Status  MsgStatus
}

// ListMsgResult 消息列表结果
type ListMsgResult struct {
	Messages []*RawSystemMsg
	ChatId   string
	HasMore  bool
}
