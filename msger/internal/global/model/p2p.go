package model

import (
	"time"
)

// 收件箱状态
type P2PInboxStatus int8

const (
	P2PInboxUnread  P2PInboxStatus = 0 // 未读
	P2PInboxRead    P2PInboxStatus = 1 // 已读
	P2PInboxRevoked P2PInboxStatus = 2 // 已撤回
)

// 参数定义
const (
	MaxTextLength = 500
	MaxRevokeTime = time.Second * 5
)
