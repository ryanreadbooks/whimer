package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/rpc"
	"github.com/ryanreadbooks/whimer/note/internal/svc"
	"github.com/ryanreadbooks/whimer/note/sdk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "etc/note.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	ctx := svc.NewServiceContext(&c)

	grpcServer := zrpc.MustNewServer(c.Grpc, func(s *grpc.Server) {
		sdk.RegisterNoteServer(s, rpc.NewNoteServer(ctx))
		if c.Grpc.Mode == service.DevMode || c.Grpc.Mode == service.TestMode {
			reflection.Register(s)
		}
	})
	grpcServer.AddUnaryInterceptors(
		interceptor.ServerErrorHandle,
		interceptor.ServerMetadataExtract,
	)
	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(grpcServer)
	logx.Info("note is serving...")
	group.Start()
}
