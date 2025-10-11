package model

import v1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"

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
	SystemMsgStatusNormal  SystemMsgStatus = SystemMsgStatus(v1.SystemMsgStatus_SystemMsgStatus_Normal)  // 正常 （未读）
	SystemMsgStatusRevoked SystemMsgStatus = SystemMsgStatus(v1.SystemMsgStatus_SystemMsgStatus_Revoked) // 被撤回
	SystemMsgStatusRead    SystemMsgStatus = SystemMsgStatus(v1.SystemMsgStatus_SystemMsgStatus_Read)    // 已读
)

type SystemNotifyMentionMsg struct {
	Uid     int64  `json:"uid"`     // @人的用户
	Target  int64  `json:"target"`  // 被@的用户
	Content []byte `json:"content"` // 被@的完整内容
}
