package dep

import (
	"github.com/ryanreadbooks/whimer/misc/idgen"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/msger/internal/config"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"

	foliumsdk "github.com/ryanreadbooks/folium/sdk"
)

var (
	userer userv1.UserServiceClient
	idGen  foliumsdk.IClient
	err    error
)

func Init(c *config.Config) {
	userer = xgrpc.NewRecoverableClient(c.External.Grpc.Passport,
		userv1.NewUserServiceClient, func(cc userv1.UserServiceClient) { userer = cc },
	)

	initIdGen(c)
}

func Userer() userv1.UserServiceClient {
	return userer
}

func Idgen() foliumsdk.IClient {
	return idGen
}

func initIdGen(c *config.Config) {
	idGen = idgen.GetIdgen(c.Seqer.Addr, func(newIdgen foliumsdk.IClient) { idGen = newIdgen })
}
