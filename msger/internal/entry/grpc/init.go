package grpc

import (
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	p2pv1 "github.com/ryanreadbooks/whimer/msger/api/p2p/v1"
	"github.com/ryanreadbooks/whimer/msger/internal/srv"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func Init(c zrpc.RpcServerConf, svc *srv.Service) *zrpc.RpcServer {
	server := zrpc.MustNewServer(c, func(s *grpc.Server) {
		p2pv1.RegisterChatServiceServer(s, NewChatServiceServer(svc))
		xgrpc.EnableReflectionIfNecessary(c, s)
	})
	interceptor.InstallUnaryServerInterceptors(server)

	return server
}
