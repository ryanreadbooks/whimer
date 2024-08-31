package comment

import (
	"sync/atomic"

	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"

	replysdk "github.com/ryanreadbooks/whimer/comment/sdk/v1"
)

var (
	// 笔记服务
	commenter replysdk.ReplyClient
	// 是否可用
	available atomic.Bool
)

func Init(c *config.Config) {
	cli, err := zrpc.NewClient(
		c.Backend.Comment.AsZrpcClientConf(),
		zrpc.WithUnaryClientInterceptor(interceptor.ClientMetadataInject))
	if err != nil {
		logx.Errorf("external init: can not init comment")
	} else {
		commenter = replysdk.NewReplyClient(cli.Conn())
		available.Store(true)
	}
}

func GetCommenter() replysdk.ReplyClient {
	return commenter
}
