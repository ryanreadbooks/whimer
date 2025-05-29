package biz

import "github.com/ryanreadbooks/whimer/msger/internal/biz/p2p"

type Biz struct {
	P2PBiz p2p.ChatBiz
}

func New() Biz {
	return Biz{
		P2PBiz: p2p.NewP2PChatBiz(),
	}
}
