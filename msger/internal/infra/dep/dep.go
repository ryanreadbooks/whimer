package dep

import (
	"github.com/ryanreadbooks/whimer/misc/idgen"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/msger/internal/config"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	wspushv1 "github.com/ryanreadbooks/whimer/wslink/api/push/v1"

	foliumsdk "github.com/ryanreadbooks/folium/sdk"
)

var (
	userer   userv1.UserServiceClient
	wsLinker wspushv1.PushServiceClient
	idGen    foliumsdk.IClient
)

func Init(c *config.Config) {
	initIdGen(c)
	userer = xgrpc.NewRecoverableClient(c.External.Grpc.Passport,
		userv1.NewUserServiceClient, func(cc userv1.UserServiceClient) { userer = cc },
	)

	wsLinker = xgrpc.NewRecoverableClient(c.External.Grpc.Wslink,
		wspushv1.NewPushServiceClient, func(psc wspushv1.PushServiceClient) { wsLinker = psc },
	)
}

func Userer() userv1.UserServiceClient {
	return userer
}

func WsLinker() wspushv1.PushServiceClient {
	return wsLinker
}

func Idgen() foliumsdk.IClient {
	return idGen
}

func initIdGen(c *config.Config) {
	idGen = idgen.GetIdgen(c.Seqer.Addr, func(newIdgen foliumsdk.IClient) { idGen = newIdgen })
}
