package grpc

import (
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor/checker"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
	"github.com/ryanreadbooks/whimer/search/internal/srv"
	"github.com/zeromicro/go-zero/zrpc"

	"google.golang.org/grpc"
)

func Init(c zrpc.RpcServerConf, svc *srv.Service) *zrpc.RpcServer {
	grpcServer := zrpc.MustNewServer(c, func(s *grpc.Server) {
		searchv1.RegisterSearchServiceServer(s, NewSearchService(svc))
		searchv1.RegisterDocumentServiceServer(s, NewDocumentService(svc))
		xgrpc.EnableReflectionIfNecessary(c, s)
	})

	interceptor.InstallUnaryServerInterceptors(grpcServer,
		interceptor.WithUnaryChecker(
			checker.UidExistenceWithOpt(
				checker.WithServicesIgnore(searchv1.SearchService_ServiceDesc.ServiceName),
				checker.WithMethodsIgnore(
					searchv1.DocumentService_BatchUpdateNoteCommentCount_FullMethodName,
					searchv1.DocumentService_BatchUpdateNoteLikeCount_FullMethodName),
			),
		),
	)

	return grpcServer
}
