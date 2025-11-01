package model

import "time"

type ChatType int8

const (
	P2PChat   ChatType = 1
	GroupChat ChatType = 2
)

type ChatStatus int8

const (
	ChatStatusNormal ChatStatus = 0
)

// 用户收件箱状态
type ChatInboxStatus int8

const (
	ChatInboxStatusNormal  ChatInboxStatus = 0
	ChatInboxStatusDeleted ChatInboxStatus = 1
)

// 参数定义
const (
	MaxTextLength = 500
	MaxRevokeTime = time.Second * 5
)
