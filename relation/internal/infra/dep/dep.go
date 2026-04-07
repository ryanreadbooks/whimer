package dep

import (
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
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
