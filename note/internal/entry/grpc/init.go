package grpc

import (
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor/checker"
	"github.com/ryanreadbooks/whimer/note/internal/srv"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
	"github.com/zeromicro/go-zero/zrpc"

	"google.golang.org/grpc"
)

func Init(c zrpc.RpcServerConf, service *srv.Service) *zrpc.RpcServer {
	grpcServer := zrpc.MustNewServer(c, func(s *grpc.Server) {
		notev1.RegisterNoteCreatorServiceServer(s, NewNoteAdminServiceServer(service))
		notev1.RegisterNoteFeedServiceServer(s, NewNoteFeedServiceServer(service))
		notev1.RegisterNoteInteractServiceServer(s, NewNoteInteractServiceServer(service))
		xgrpc.EnableReflectionIfNecessary(c, s)
	})

	interceptor.InstallUnaryServerInterceptors(grpcServer,
		interceptor.WithUnaryChecker(
			checker.UidExistenceWithOpt(
				checker.WithIgnore(NoteFeedServiceName),
			),
		),
	)

	return grpcServer
}