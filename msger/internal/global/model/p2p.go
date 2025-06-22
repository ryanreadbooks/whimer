package model

import (
	"time"

	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
)

// 定义消息的类型
// 在pb中定义
type MsgType = pbmsg.MsgType

const (
	MsgText  MsgType = pbmsg.MsgType_MSG_TYPE_TEXT
	MsgImage MsgType = pbmsg.MsgType_MSG_TYPE_IMAGE
	MsgVideo MsgType = pbmsg.MsgType_MSG_TYPE_VIDEO
)

// 定义消息的状态
type MsgStatus = pbmsg.MsgStatus

const (
	MsgStatusNormal  MsgStatus = pbmsg.MsgStatus_MSG_STATUS_NORMAL // 正常
	MsgStatusRevoked MsgStatus = pbmsg.MsgStatus_MSG_STATUS_REVOKE // 撤回
)

// 收件箱状态
type InboxStatus int8

const (
	InboxUnread  InboxStatus = 0 // 未读
	InboxRead    InboxStatus = 1 // 已读
	InboxRevoked InboxStatus = 2 // 已撤回
)

// 参数定义
const (
	MaxTextLength = 500
	MaxRevokeTime = time.Second * 5
)
