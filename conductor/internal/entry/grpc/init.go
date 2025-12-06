package grpc

import (
	taskservice "github.com/ryanreadbooks/whimer/conductor/api/taskservice/v1"
	"github.com/ryanreadbooks/whimer/conductor/internal/service"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func Init(c zrpc.RpcServerConf, srv *service.Service) *zrpc.RpcServer {
	server := zrpc.MustNewServer(c, func(s *grpc.Server) {
		taskservice.RegisterTaskServiceServer(s, NewTaskServiceServer(srv))
		xgrpc.EnableReflectionIfNecessary(c, s)
	})

	return server
}
