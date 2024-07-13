package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/passport/internal/config"
	"github.com/ryanreadbooks/whimer/passport/internal/handler"
	stubaccess "github.com/ryanreadbooks/whimer/passport/internal/stub/access"
	"github.com/ryanreadbooks/whimer/passport/internal/svc"
	"github.com/ryanreadbooks/whimer/passport/sdk/access"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/passport.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	ctx := svc.NewServiceContext(&c)
	restServer := rest.MustNewServer(c.Http)

	grpcServer := zrpc.MustNewServer(c.Grpc, func(s *grpc.Server) {
		access.RegisterAccessServer(s, stubaccess.NewAccessServer(ctx))
		if c.Grpc.Mode == service.DevMode || c.Grpc.Mode == service.TestMode {
			reflection.Register(s)
		}
	})

	handler.RegisterHandlers(restServer, ctx)
	logx.Info("passport is serving...")
	group := service.NewServiceGroup()
	group.Add(restServer)
	group.Add(grpcServer)
	group.Start()
}
