package infra

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	msgv1 "github.com/ryanreadbooks/whimer/msger/api/p2p/v1"
	systemv1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"

	"github.com/ryanreadbooks/whimer/misc/xgrpc"
)

// 消息服务
var (
	chatter        msgv1.ChatServiceClient
	systemNotifier systemv1.NotificationServiceClient
	systemChatter  systemv1.ChatServiceClient
)

func InitMsger(c *config.Config) {
	chatter = xgrpc.NewRecoverableClient(c.Backend.Msger,
		msgv1.NewChatServiceClient,
		func(cc msgv1.ChatServiceClient) {
			chatter = cc
		})

	systemNotifier = xgrpc.NewRecoverableClient(c.Backend.Msger,
		systemv1.NewNotificationServiceClient,
		func(cc systemv1.NotificationServiceClient) {
			systemNotifier = cc
		})

	systemChatter = xgrpc.NewRecoverableClient(c.Backend.Msger,
		systemv1.NewChatServiceClient,
		func(cc systemv1.ChatServiceClient) {
			systemChatter = cc
		})
}

func P2PChatter() msgv1.ChatServiceClient {
	return chatter
}

func SystemNotifier() systemv1.NotificationServiceClient {
	return systemNotifier
}

func SystemChatter() systemv1.ChatServiceClient {
	return systemChatter
}
