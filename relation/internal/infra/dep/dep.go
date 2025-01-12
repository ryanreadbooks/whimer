package dep

import (
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	"github.com/ryanreadbooks/whimer/relation/internal/config"
)

var (
	userer userv1.UserServiceClient

	err error
)

func Init(c *config.Config) {
	userer = xgrpc.NewRecoverableClient(c.Backend.Passport,
		userv1.NewUserServiceClient, func(nc userv1.UserServiceClient) { userer = nc })
}

func Userer() userv1.UserServiceClient {
	return userer
}
