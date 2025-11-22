package biz

import (
	"github.com/ryanreadbooks/whimer/msger/internal/biz/system"
	"github.com/ryanreadbooks/whimer/msger/internal/biz/userchat"
)

type Biz struct {
	SystemBiz system.ChatBiz

	ChatBiz       userchat.ChatBiz
	ChatMemberBiz userchat.ChatMemberBiz
	MsgBiz        userchat.MsgBiz
	ChatInboxBiz  userchat.ChatInboxBiz
}

func New() Biz {
	return Biz{
		SystemBiz: system.NewSystemChatBiz(),

		ChatBiz:       userchat.NewChatBiz(),
		ChatMemberBiz: userchat.NewChatMemberBiz(),
		MsgBiz:        userchat.NewMsgBiz(),
		ChatInboxBiz:  userchat.NewChatInboxBiz(),
	}
}
