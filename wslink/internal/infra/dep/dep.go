package dep

import (
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	"github.com/ryanreadbooks/whimer/passport/pkg/middleware/auth"
	"github.com/ryanreadbooks/whimer/wslink/internal/config"
)

var (
	auther *auth.Auth
	userer userv1.UserServiceClient
)

func Init(c *config.Config) {
	auther = auth.RecoverableAuther(c.Backend.Passport)
	userer = xgrpc.NewRecoverableClient(c.Backend.Passport,
		userv1.NewUserServiceClient, func(nc userv1.UserServiceClient) { userer = nc })
}

func Userer() userv1.UserServiceClient {
	return userer
}

func Auther() *auth.Auth {
	return auther
}
