package grpc

import (
	namespaceservice "github.com/ryanreadbooks/whimer/conductor/api/namespaceservice/v1"
	taskservice "github.com/ryanreadbooks/whimer/conductor/api/taskservice/v1"
	workerservice "github.com/ryanreadbooks/whimer/conductor/api/workerservice/v1"
	"github.com/ryanreadbooks/whimer/conductor/internal/service"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func Init(c zrpc.RpcServerConf, srv *service.Service) *zrpc.RpcServer {
	server := zrpc.MustNewServer(c, func(s *grpc.Server) {
		namespaceservice.RegisterNamespaceServiceServer(s, NewNamespaceServiceServer(srv))
		taskservice.RegisterTaskServiceServer(s, NewTaskServiceServer(srv))
		workerservice.RegisterWorkerServiceServer(s, NewWorkerServiceServer(srv))
		xgrpc.EnableReflectionIfNecessary(c, s)
	})

	interceptor.InstallUnaryServerInterceptors(server)

	return server
}
