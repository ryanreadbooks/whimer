package model

import (
	"github.com/ryanreadbooks/whimer/misc/xerror"
	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
)

// 定义消息的类型
type MsgType int16

const (
	MsgTypeUnknown MsgType = 0
	MsgText        MsgType = 1  // 纯文本
	MsgImage       MsgType = 10 // 纯图片
	MsgVideo       MsgType = 20 // 视频
)

func MsgTypeToPb(t MsgType) pbmsg.MsgType {
	return pbmsg.MsgType(t)
}

func MsgTypeFromPb(t pbmsg.MsgType) (MsgType, error) {
	switch t {
	case pbmsg.MsgType_MSG_TYPE_TEXT:
		return MsgText, nil
	case pbmsg.MsgType_MSG_TYPE_IMAGE:
		return MsgImage, nil
	case pbmsg.MsgType_MSG_TYPE_VIDEO:
		return MsgVideo, nil
	default:
		return 0, xerror.ErrArgs.Msg("unsupported msg type")
	}
}

// 定义消息的状态
type MsgStatus int8

const (
	MsgStatusNormal  MsgStatus = 1 // 正常
	MsgStatusRevoked MsgStatus = 2 // 撤回
)

func MsgStatusToPb(s MsgStatus) pbmsg.MsgStatus {
	return pbmsg.MsgStatus(s)
}

// 收信箱中消息状态
type InboxMsgStatus int8

const (
	InboxMsgStatusNormal  InboxMsgStatus = 1 // 正常（未读）
	InboxMsgStatusRevoked InboxMsgStatus = 2 // 撤回和MsgStatus保持一致
	InboxMsgStatusRead    InboxMsgStatus = 3 // 已读
)
