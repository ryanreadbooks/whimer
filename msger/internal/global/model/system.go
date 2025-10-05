package model

// 系统消息

// 系统绘画类型
type SystemChatType int8

const (
	SystemNotificationChat      SystemChatType = 1
	SystemReplyToMeChat         SystemChatType = 2
	SystemMentionedByOthersChat SystemChatType = 3
	SystemLikeReceivedChat      SystemChatType = 4
)

func (s SystemChatType) Desc() string {
	switch s {
	case SystemNotificationChat:
		return "系统通知"
	case SystemReplyToMeChat:
		return "回复我的"
	case SystemMentionedByOthersChat:
		return "@我的"
	case SystemLikeReceivedChat:
		return "收到的赞"
	}
	return "通知"
}

// 系统消息类型
type SystemMsgStatus int8

const (
	SystemMsgStatusNormal  = 1 // 正常
	SystemMsgStatusRevoked = 2 // 被撤回
	SystemMsgStatusRead    = 3 // 已读
)
