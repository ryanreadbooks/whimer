package grpc

import (
	"github.com/ryanreadbooks/whimer/counter/internal/srv"
	v1 "github.com/ryanreadbooks/whimer/idl/gen/go/counter/api/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor/checker"
	"github.com/zeromicro/go-zero/zrpc"

	"google.golang.org/grpc"
)

func Init(c zrpc.RpcServerConf, svc *srv.Service) *zrpc.RpcServer {
	server := zrpc.MustNewServer(c, func(s *grpc.Server) {
		v1.RegisterCounterServiceServer(s, NewCounterServer(svc))
		xgrpc.EnableReflectionIfNecessary(c, s)
	})
	interceptor.InstallUnaryServerInterceptors(server,
		interceptor.WithUnaryChecker(checker.UidExistence),
	)

	return server
}
