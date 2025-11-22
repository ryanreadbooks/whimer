package userchat

import (
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/msger/internal/biz/userchat"
)

// Msg + chatId
type ChatMsg struct {
	*userchat.Msg
	ChatId uuid.UUID `json:"chat_id"`
	Pos    int64     `json:"pos"`
}

func makeChatMsgFromMsg(m *userchat.Msg) *ChatMsg {
	return &ChatMsg{
		Msg: m,
	}
}
