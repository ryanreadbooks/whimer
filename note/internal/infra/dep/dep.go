package dep

import (
	"github.com/ryanreadbooks/whimer/conductor/pkg/go/sdk/producer"
	commentv1 "github.com/ryanreadbooks/whimer/idl/gen/go/comment/api/v1"
	counterv1 "github.com/ryanreadbooks/whimer/idl/gen/go/counter/api/v1"
	userv1 "github.com/ryanreadbooks/whimer/idl/gen/go/passport/api/user/v1"
	searchv1 "github.com/ryanreadbooks/whimer/idl/gen/go/search/api/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/note/internal/config"
)

var (
	counter         counterv1.CounterServiceClient // 计数服务
	commenter       commentv1.CommentServiceClient // 评论服务
	searchdocer     searchv1.DocumentServiceClient // 搜索服务
	userer          userv1.UserServiceClient
	conductProducer *producer.Client
)

func Init(c *config.Config) {
	counter = counterv1.NewCounterServiceClient(
		xgrpc.NewRecoverableClientConn(c.External.Grpc.Counter),
	)
	commenter = commentv1.NewCommentServiceClient(
		xgrpc.NewRecoverableClientConn(c.External.Grpc.Comment),
	)
	searchdocer = searchv1.NewDocumentServiceClient(
		xgrpc.NewRecoverableClientConn(c.External.Grpc.Search),
	)
	userer = userv1.NewUserServiceClient(
		xgrpc.NewRecoverableClientConn(c.External.Grpc.Passport),
	)

	conductProducer, _ = producer.New(producer.ClientOptions{
		HostConf:  c.External.Grpc.Conductor,
		Namespace: "note",
	})
}

func GetCounter() counterv1.CounterServiceClient {
	return counter
}

func GetCommenter() commentv1.CommentServiceClient {
	return commenter
}

func GetSearchDocer() searchv1.DocumentServiceClient {
	return searchdocer
}

func GetUserer() userv1.UserServiceClient {
	return userer
}

func GetConductProducer() *producer.Client {
	return conductProducer
}
