package vo

type SystemMessage struct {
	Uid       int64  // 发送者
	TargetUid int64  // 接受者
	Content   []byte // 内容
}
