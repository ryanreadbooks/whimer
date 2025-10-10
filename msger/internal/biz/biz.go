package biz

import (
	"github.com/ryanreadbooks/whimer/msger/internal/biz/p2p"
	"github.com/ryanreadbooks/whimer/msger/internal/biz/system"
	"github.com/ryanreadbooks/whimer/msger/internal/biz/websocket"
)

type Biz struct {
	P2PBiz      p2p.ChatBiz
	SystemBiz   system.ChatBiz
	WebsocketBiz websocket.Biz
}

func New() Biz {
	return Biz{
		P2PBiz:          p2p.NewP2PChatBiz(),
		SystemBiz:       system.NewSystemChatBiz(),
		WebsocketBiz:     websocket.NewWebsocketBiz(),
	}
}
