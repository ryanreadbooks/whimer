package dep

import (
	countersdk "github.com/ryanreadbooks/whimer/counter/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/passport/sdk/middleware/auth"
	"github.com/zeromicro/go-zero/core/logx"
)

var (
	auther *auth.Auth
	// 计数服务
	counter countersdk.CounterServiceClient
	err     error
)

func Init(c *config.Config) {
	auther = auth.MustAuther(c.External.Grpc.Passport)

	counter, err = xgrpc.NewClient(c.External.Grpc.Counter,
		countersdk.NewCounterServiceClient)
	if err != nil {
		logx.Errorf("external init: can not init counter: %v", err)
	}
}

func GetAuther() *auth.Auth {
	return auther
}

func GetCounter() countersdk.CounterServiceClient {
	return counter
}
