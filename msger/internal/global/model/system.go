package model

// 系统消息

// 系统会话类型
type SystemChatType int8

const (
	SystemNotifyNoticeChat    SystemChatType = 1
	SystemNotifyReplyChat     SystemChatType = 2
	SystemNotifyMentionedChat SystemChatType = 3
	SystemNotifyLikesChat     SystemChatType = 4
)

func (s SystemChatType) Desc() string {
	switch s {
	case SystemNotifyNoticeChat:
		return "系统通知"
	case SystemNotifyReplyChat:
		return "回复我的"
	case SystemNotifyMentionedChat:
		return "@我的"
	case SystemNotifyLikesChat:
		return "收到的赞"
	}
	return "通知"
}

// 系统消息类型
type SystemMsgStatus int8

const (
	SystemMsgStatusNormal  = 1 // 正常 （未读）
	SystemMsgStatusRevoked = 2 // 被撤回
	SystemMsgStatusRead    = 3 // 已读
)
