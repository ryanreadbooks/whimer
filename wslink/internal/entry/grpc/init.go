package grpc

import (
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"

	forwardv1 "github.com/ryanreadbooks/whimer/wslink/api/forward/v1"
	pushv1 "github.com/ryanreadbooks/whimer/wslink/api/push/v1"
	"github.com/ryanreadbooks/whimer/wslink/internal/srv"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func Init(c zrpc.RpcServerConf, svc *srv.Service) *zrpc.RpcServer {
	server := zrpc.MustNewServer(c, func(s *grpc.Server) {
		pushv1.RegisterPushServiceServer(s, NewPushServiceServer(svc))
		forwardv1.RegisterForwardServiceServer(s, NewForwardServiceServer(svc))
		xgrpc.EnableReflectionIfNecessary(c, s)
	})

	interceptor.InstallUnaryServerInterceptors(server)

	return server
}
