package dep

import (
	userv1 "github.com/ryanreadbooks/whimer/idl/gen/go/passport/api/user/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/passport/pkg/middleware/auth"
	"github.com/ryanreadbooks/whimer/wslink/internal/config"
)

var (
	auther *auth.Auth
	userer userv1.UserServiceClient
)

func Init(c *config.Config) {
	auther = auth.RecoverableAuther(c.Backend.Passport)
	userer = userv1.NewUserServiceClient(
		xgrpc.NewRecoverableClientConn(c.Backend.Passport),
	)
}

func Userer() userv1.UserServiceClient {
	return userer
}

func Auther() *auth.Auth {
	return auther
}
