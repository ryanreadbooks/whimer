package dep

import (
	systemv1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
	userchatv1 "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"

	"github.com/ryanreadbooks/whimer/misc/xgrpc"
)

// 消息服务
var (
	systemNotifier systemv1.NotificationServiceClient
	systemChatter  systemv1.ChatServiceClient

	userChatter userchatv1.UserChatServiceClient
)

func InitMsger(c *config.Config) {
	conn := xgrpc.NewRecoverableClientConn(c.Backend.Msger)
	systemNotifier = systemv1.NewNotificationServiceClient(conn)
	systemChatter = systemv1.NewChatServiceClient(conn)
	userChatter = userchatv1.NewUserChatServiceClient(conn)
}

func SystemNotifier() systemv1.NotificationServiceClient {
	return systemNotifier
}

func SystemChatter() systemv1.ChatServiceClient {
	return systemChatter
}

func UserChatter() userchatv1.UserChatServiceClient {
	return userChatter
}
