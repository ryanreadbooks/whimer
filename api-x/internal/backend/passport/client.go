package passport

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/ryanreadbooks/whimer/passport/sdk/middleware/auth"
	user "github.com/ryanreadbooks/whimer/passport/sdk/user/v1"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

var (
	auther *auth.Auth
	userer user.UserClient
	err    error
)

func Init(c *config.Config) {
	auther = auth.MustAuther(c.Backend.Passport)

	cli, err := zrpc.NewClient(
		c.Backend.Passport.AsZrpcClientConf(),
		zrpc.WithUnaryClientInterceptor(interceptor.UnaryClientMetadataInject))
	if err != nil {
		logx.Errorf("external init: can not init passport user")
	} else {
		userer = user.NewUserClient(cli.Conn())
	}
}

func Auther() *auth.Auth {
	return auther
}

func Userer() user.UserClient {
	return userer
}
