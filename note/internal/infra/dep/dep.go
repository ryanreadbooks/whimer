package dep

import (
	commentv1 "github.com/ryanreadbooks/whimer/comment/sdk/v1"
	counterv1 "github.com/ryanreadbooks/whimer/counter/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/note/internal/config"
)

var (
	counter   counterv1.CounterServiceClient // 计数服务
	commenter commentv1.ReplyServiceClient   // 评论服务
)

func Init(c *config.Config) {
	counter = xgrpc.NewRecoverableClient(c.External.Grpc.Counter,
		counterv1.NewCounterServiceClient,
		func(nc counterv1.CounterServiceClient) { counter = nc })

	commenter = xgrpc.NewRecoverableClient(c.External.Grpc.Comment,
		commentv1.NewReplyServiceClient, func(nc commentv1.ReplyServiceClient) { commenter = nc })
}

func GetCounter() counterv1.CounterServiceClient {
	return counter
}

func GetCommenter() commentv1.ReplyServiceClient {
	return commenter
}
