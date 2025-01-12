package grpc

import (
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"
	"github.com/ryanreadbooks/whimer/passport/internal/srv"
	accessv1 "github.com/ryanreadbooks/whimer/passport/api/access/v1"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func Init(c zrpc.RpcServerConf, service *srv.Service) *zrpc.RpcServer {
	grpcServer := zrpc.MustNewServer(c, func(s *grpc.Server) {
		accessv1.RegisterAccessServiceServer(s, NewAccessServiceServer(service))
		userv1.RegisterUserServiceServer(s, NewUserServiceServer(service))
		xgrpc.EnableReflectionIfNecessary(c, s)
	})

	interceptor.InstallUnaryServerInterceptors(grpcServer)

	return grpcServer
}
