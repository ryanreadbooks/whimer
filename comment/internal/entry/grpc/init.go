package grpc

import (
	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/comment/internal/srv"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor/checker"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func Init(c zrpc.RpcServerConf, svc *srv.Service) *zrpc.RpcServer {
	server := zrpc.MustNewServer(c, func(s *grpc.Server) {
		commentv1.RegisterReplyServiceServer(s, NewReplyServiceServer(svc))
		xgrpc.EnableReflectionIfNecessary(c, s)
	})
	interceptor.InstallUnaryServerInterceptors(server,
		interceptor.WithUnaryChecker(
			checker.UidExistenceWithOpt(
				checker.WithMethodsIgnore(uidCheckIgnoredMethods...),
			),
		))

	return server
}
