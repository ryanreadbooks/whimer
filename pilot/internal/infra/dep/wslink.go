package dep

import (
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	pushv1 "github.com/ryanreadbooks/whimer/wslink/api/push/v1"
)

var (
	wsLinker pushv1.PushServiceClient
)

func InitWsLink(c *config.Config) {
	wsLinker = xgrpc.NewRecoverableClient(c.Backend.WsLink,
		pushv1.NewPushServiceClient,
		func(cc pushv1.PushServiceClient) { wsLinker = cc })
}

func WebsocketPusher() pushv1.PushServiceClient {
	return wsLinker
}
