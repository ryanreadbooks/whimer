package dep

import (
	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	counterv1 "github.com/ryanreadbooks/whimer/counter/api/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
)

var (
	counter     counterv1.CounterServiceClient // 计数服务
	commenter   commentv1.ReplyServiceClient   // 评论服务
	searchdocer searchv1.DocumentServiceClient // 搜索服务
	userer      userv1.UserServiceClient
)

func Init(c *config.Config) {
	counter = xgrpc.NewRecoverableClient(c.External.Grpc.Counter,
		counterv1.NewCounterServiceClient,
		func(nc counterv1.CounterServiceClient) { counter = nc })

	commenter = xgrpc.NewRecoverableClient(c.External.Grpc.Comment,
		commentv1.NewReplyServiceClient, func(nc commentv1.ReplyServiceClient) { commenter = nc })

	searchdocer = xgrpc.NewRecoverableClient(c.External.Grpc.Search,
		searchv1.NewDocumentServiceClient, func(nc searchv1.DocumentServiceClient) { searchdocer = nc })

	userer = xgrpc.NewRecoverableClient(c.External.Grpc.Passport,
		userv1.NewUserServiceClient, func(cc userv1.UserServiceClient) { userer = cc })

}

func GetCounter() counterv1.CounterServiceClient {
	return counter
}

func GetCommenter() commentv1.ReplyServiceClient {
	return commenter
}

func GetSearchDocer() searchv1.DocumentServiceClient {
	return searchdocer
}

func GetUserer() userv1.UserServiceClient {
	return userer
}
