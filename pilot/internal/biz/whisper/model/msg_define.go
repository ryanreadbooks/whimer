package model

import (
	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
)

type MsgType int32

const (
	MsgTypeUnspecified MsgType = MsgType(pbmsg.MsgStatus_MSG_STATUS_UNSPECIFIED)
	MsgText            MsgType = MsgType(pbmsg.MsgType_MSG_TYPE_TEXT)
	MsgImage           MsgType = MsgType(pbmsg.MsgType_MSG_TYPE_IMAGE)
)

var (
	validMsgTypeMap = map[MsgType]struct{}{
		MsgText:  {},
		MsgImage: {},
	}
)

func IsValidMsgType(t MsgType) bool {
	_, ok := validMsgTypeMap[t]
	return ok
}

type MsgStatus int32

const (
	MsgStatusNormal MsgStatus = MsgStatus(pbmsg.MsgStatus_MSG_STATUS_NORMAL)
	MsgStatusRecall MsgStatus = MsgStatus(pbmsg.MsgStatus_MSG_STATUS_RECALL)
)

const (
	MaxTextContentLength = 500
)
