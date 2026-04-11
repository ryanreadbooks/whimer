package dep

import (
	userv1 "github.com/ryanreadbooks/whimer/idl/gen/go/passport/api/user/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/relation/internal/config"
)

var (
	userer userv1.UserServiceClient
)

func Init(c *config.Config) {
	userer = userv1.NewUserServiceClient(
		xgrpc.NewRecoverableClientConn(c.Backend.Passport),
	)
}

func Userer() userv1.UserServiceClient {
	return userer
}
