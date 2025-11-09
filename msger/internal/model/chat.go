package model

import (
	"time"

	pbuserchat "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
)

// 会话类型
type ChatType int8

const (
	P2PChat   ChatType = 1
	GroupChat ChatType = 2
)

func ChatTypeFromPb(b pbuserchat.ChatType) (ChatType, error) {
	switch b {
	case pbuserchat.ChatType_P2P:
		return P2PChat, nil
	case pbuserchat.ChatType_GROUP:
		return GroupChat, nil
	default:
		return 0, global.ErrArgs.Msg("unsupported chat type")
	}
}

func ChatTypeToPb(c ChatType) pbuserchat.ChatType {
	switch c {
	case P2PChat:
		return pbuserchat.ChatType_P2P
	case GroupChat:
		return pbuserchat.ChatType_GROUP
	default:
		return pbuserchat.ChatType_CHAT_TYPE_UNSPECIFIED
	}
}

// 会话状态
type ChatStatus int8

const (
	ChatStatusNormal ChatStatus = 1
)

func ChatStatusFromPb(b pbuserchat.ChatStatus) (ChatStatus, error) {
	switch b {
	case pbuserchat.ChatStatus_NORMAL:
		return ChatStatusNormal, nil
	default:
		return 0, global.ErrArgs.Msg("unsupported chat status")
	}
}

func ChatStatusToPb(s ChatStatus) pbuserchat.ChatStatus {
	switch s {
	case ChatStatusNormal:
		return pbuserchat.ChatStatus_NORMAL
	default:
		return pbuserchat.ChatStatus_CHAT_STATUS_UNSPECIFIED
	}
}

// 用户收件箱状态
type ChatInboxStatus int8

const (
	ChatInboxStatusNormal  ChatInboxStatus = 0
	ChatInboxStatusDeleted ChatInboxStatus = 1
)

// 参数定义
const (
	MaxTextLength = 500
	MaxRecallTime = time.Second * 5
)

type ChatInboxPinState int8

const (
	ChatInboxUnPinned ChatInboxPinState = 0
	ChatInboxPinned   ChatInboxPinState = 1
)
