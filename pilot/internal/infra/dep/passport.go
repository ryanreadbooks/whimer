package dep

import (
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	"github.com/ryanreadbooks/whimer/passport/pkg/middleware/auth"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
)

var (
	auther *auth.Auth
	userer userv1.UserServiceClient
)

func InitPassport(c *config.Config) {
	auther = auth.MustAuther(c.Backend.Passport)
	userer = xgrpc.NewRecoverableClient(c.Backend.Passport, userv1.NewUserServiceClient, func(cc userv1.UserServiceClient) { userer = cc })
}

func Auther() *auth.Auth {
	return auther
}

func Userer() userv1.UserServiceClient {
	return userer
}
