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
	auther = auth.RecoverableAuther(c.Backend.Passport)
	userer = userv1.NewUserServiceClient(
		xgrpc.NewRecoverableClientConn(c.Backend.Passport),
	)
}

func Auther() *auth.Auth {
	return auther
}

func Userer() userv1.UserServiceClient {
	return userer
}
