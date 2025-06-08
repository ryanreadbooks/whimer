package dep

import (
	"github.com/ryanreadbooks/whimer/misc/idgen"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/msger/internal/config"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	pushv1 "github.com/ryanreadbooks/whimer/wslink/api/push/v1"

	foliumsdk "github.com/ryanreadbooks/folium/sdk"
)

var (
	userer   userv1.UserServiceClient
	notifier pushv1.PushServiceClient
	idGen    foliumsdk.IClient
	err      error
)

func Init(c *config.Config) {
	initIdGen(c)
	userer = xgrpc.NewRecoverableClient(c.External.Grpc.Passport,
		userv1.NewUserServiceClient, func(cc userv1.UserServiceClient) { userer = cc },
	)

	notifier = xgrpc.NewRecoverableClient(c.External.Grpc.Wslink,
		pushv1.NewPushServiceClient, func(psc pushv1.PushServiceClient) { notifier = psc },
	)
}

func Userer() userv1.UserServiceClient {
	return userer
}

func Notifier() pushv1.PushServiceClient {
	return notifier
}

func Idgen() foliumsdk.IClient {
	return idGen
}

func initIdGen(c *config.Config) {
	idGen = idgen.GetIdgen(c.Seqer.Addr, func(newIdgen foliumsdk.IClient) { idGen = newIdgen })
}
