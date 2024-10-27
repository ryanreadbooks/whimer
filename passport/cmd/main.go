package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/ryanreadbooks/whimer/passport/internal/config"
	"github.com/ryanreadbooks/whimer/passport/internal/handler"
	accrpc "github.com/ryanreadbooks/whimer/passport/internal/rpc/access"
	userrpc "github.com/ryanreadbooks/whimer/passport/internal/rpc/user"
	"github.com/ryanreadbooks/whimer/passport/internal/svc"
	access "github.com/ryanreadbooks/whimer/passport/sdk/access/v1"
	user "github.com/ryanreadbooks/whimer/passport/sdk/user/v1"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

var configFile = flag.String("f", "etc/passport.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	ctx := svc.NewServiceContext(&c)
	restServer := rest.MustNewServer(c.Http)
	handler.RegisterHandlers(restServer, ctx)

	grpcServer := zrpc.MustNewServer(c.Grpc, func(s *grpc.Server) {
		// 访问认证
		access.RegisterAccessServer(s, accrpc.NewAccessServer(ctx))
		// 用户信息
		user.RegisterUserServer(s, userrpc.NewUserServer(ctx))

		// for debugging
		xgrpc.EnableReflectionIfNecessary(c.Grpc, s)
	})
	interceptor.InstallUnaryServerInterceptors(grpcServer)

	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(restServer)
	group.Add(grpcServer)

	logx.Info("passport is serving...")
	group.Start()
}
