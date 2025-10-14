package biz

import (
	"github.com/ryanreadbooks/whimer/msger/internal/biz/p2p"
	"github.com/ryanreadbooks/whimer/msger/internal/biz/system"
)

type Biz struct {
	P2PBiz    p2p.ChatBiz
	SystemBiz system.ChatBiz
}

func New() Biz {
	return Biz{
		P2PBiz:    p2p.NewP2PChatBiz(),
		SystemBiz: system.NewSystemChatBiz(),
	}
}
