package grpc

import (
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor/checker"
	"github.com/ryanreadbooks/whimer/relation/internal/srv"
	relationv1 "github.com/ryanreadbooks/whimer/relation/sdk/v1"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func Init(c zrpc.RpcServerConf, service *srv.Service) *zrpc.RpcServer {
	grpcServer := zrpc.MustNewServer(c, func(s *grpc.Server) {
		relationv1.RegisterRelationServiceServer(s, NewRelationServiceServer(service))
		xgrpc.EnableReflectionIfNecessary(c, s)
	})

	interceptor.InstallUnaryServerInterceptors(grpcServer,
		interceptor.UnaryServerMetadataCheck(
			checker.UidExistenceWithOpt(
				checker.WithMethodsIgnore(uidCheckIgnoredMethods...),
			),
		))

	return grpcServer
}
