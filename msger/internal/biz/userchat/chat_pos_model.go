package userchat

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dao/chat"
)

type ChatPos struct {
	ChatId uuid.UUID
	MsgId  uuid.UUID
	Pos    int64
	Ctime  int64
}

func makeChatPosFromPO(po *chat.ChatMsgPO) *ChatPos {
	return &ChatPos{
		ChatId: po.ChatId,
		MsgId:  po.MsgId,
		Pos:    po.Pos,
		Ctime:  po.Ctime,
	}
}
