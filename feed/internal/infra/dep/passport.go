package dep

import (
	"github.com/ryanreadbooks/whimer/feed/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/passport/sdk/middleware/auth"
	user "github.com/ryanreadbooks/whimer/passport/sdk/user/v1"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	auther *auth.Auth
	userer user.UserServiceClient
	err    error
)

func initPassport(c *config.Config) {
	auther = auth.MustAuther(c.Backend.Passport)

	conn, err := xgrpc.NewClientConn(c.Backend.Passport)
	if err != nil {
		logx.Errorf("dep init: can not init passport user")
	} else {
		userer = user.NewUserServiceClient(conn)
	}
}

func Auther() *auth.Auth {
	return auther
}

func Userer() user.UserServiceClient {
	return userer
}
