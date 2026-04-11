package dep

import (
	pushv1 "github.com/ryanreadbooks/whimer/idl/gen/go/wslink/api/push/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
)

var (
	wsLinker pushv1.PushServiceClient
)

func InitWsLink(c *config.Config) {
	wsLinker = pushv1.NewPushServiceClient(
		xgrpc.NewRecoverableClientConn(c.Backend.WsLink),
	)
}

func WebsocketPusher() pushv1.PushServiceClient {
	return wsLinker
}
