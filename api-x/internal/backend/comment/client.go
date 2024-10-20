package comment

import (
	"sync/atomic"

	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"

	commentv1 "github.com/ryanreadbooks/whimer/comment/sdk/v1"
)

var (
	// 笔记服务
	commenter commentv1.ReplyServiceClient
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
		commenter = commentv1.NewReplyServiceClient(cli.Conn())
		available.Store(true)
	}
}

func GetCommenter() commentv1.ReplyServiceClient {
	return commenter
}
