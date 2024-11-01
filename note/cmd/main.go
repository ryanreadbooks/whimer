package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor/checker"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/rpc"
	"github.com/ryanreadbooks/whimer/note/internal/svc"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
	"google.golang.org/grpc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "etc/note.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())
	ctx := svc.NewServiceContext(&config.Conf)

	grpcServer := zrpc.MustNewServer(config.Conf.Grpc, func(s *grpc.Server) {
		notev1.RegisterNoteAdminServiceServer(s, rpc.NewNoteAdminServiceServer(ctx))
		notev1.RegisterNoteFeedServiceServer(s, rpc.NewNoteFeedServiceServer(ctx))
		notev1.RegisterNoteInteractServiceServer(s, rpc.NewNoteInteractServiceServer(ctx))
		xgrpc.EnableReflectionIfNecessary(config.Conf.Grpc, s)
	})
	interceptor.InstallUnaryServerInterceptors(grpcServer,
		interceptor.WithUnaryChecker(
			checker.UidExistenceWithOpt(
				checker.WithIgnore(rpc.NoteFeedServiceName),
			),
		),
	)

	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(grpcServer)
	logx.Info("note is serving...")
	group.Start()
}
