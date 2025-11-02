package dep

import (
	systemv1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"

	"github.com/ryanreadbooks/whimer/misc/xgrpc"
)

// 消息服务
var (
	systemNotifier systemv1.NotificationServiceClient
	systemChatter  systemv1.ChatServiceClient
)

func InitMsger(c *config.Config) {
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

func SystemNotifier() systemv1.NotificationServiceClient {
	return systemNotifier
}

func SystemChatter() systemv1.ChatServiceClient {
	return systemChatter
}
