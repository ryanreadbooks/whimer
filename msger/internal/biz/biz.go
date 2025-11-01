package biz

import (
	"github.com/ryanreadbooks/whimer/msger/internal/biz/p2p"
	"github.com/ryanreadbooks/whimer/msger/internal/biz/system"
	"github.com/ryanreadbooks/whimer/msger/internal/biz/userchat"
)

type Biz struct {
	P2PBiz    p2p.ChatBiz
	SystemBiz system.ChatBiz

	ChatBiz       userchat.ChatBiz
	ChatMemberBiz userchat.ChatMemberBiz
	MsgBiz        userchat.MsgBiz
	ChatInboxBiz  userchat.ChatInboxBiz
}

func New() Biz {
	return Biz{
		P2PBiz:    p2p.NewP2PChatBiz(),
		SystemBiz: system.NewSystemChatBiz(),

		ChatBiz:       userchat.NewChatBiz(),
		ChatMemberBiz: userchat.NewChatMemberBiz(),
		MsgBiz:        userchat.NewMsgBiz(),
		ChatInboxBiz:  userchat.NewChatInboxBiz(),
	}
}
