package grpc

import (
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor/checker"
	p2pv1 "github.com/ryanreadbooks/whimer/msger/api/p2p/v1"
	systemv1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
	pbuserchat "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	"github.com/ryanreadbooks/whimer/msger/internal/srv"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

func Init(c zrpc.RpcServerConf, service *srv.Service) *zrpc.RpcServer {
	server := zrpc.MustNewServer(c, func(s *grpc.Server) {
		p2pv1.RegisterChatServiceServer(s, NewP2PChatServiceServer(service))
		systemv1.RegisterNotificationServiceServer(s, NewSystemNotificationServiceServer(service))
		systemv1.RegisterChatServiceServer(s, NewSystemChatServiceServer(service))
		pbuserchat.RegisterUserChatServiceServer(s, NewUserChatServiceServer(service))
		xgrpc.EnableReflectionIfNecessary(c, s)
	})
	interceptor.InstallUnaryServerInterceptors(server,
		interceptor.WithUnaryChecker(
			checker.UidExistenceWithOpt(),
		),
	)

	return server
}
