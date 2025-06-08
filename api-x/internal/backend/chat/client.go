package chat

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	msgv1 "github.com/ryanreadbooks/whimer/msger/api/p2p/v1"

	"github.com/ryanreadbooks/whimer/misc/xgrpc"
)

// 消息服务
var (
	chatter msgv1.ChatServiceClient
)

func Init(c *config.Config) {
	chatter = xgrpc.NewRecoverableClient(c.Backend.Msger,
		msgv1.NewChatServiceClient,
		func(cc msgv1.ChatServiceClient) {
			chatter = cc
		})
}
