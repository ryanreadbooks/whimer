package main

import (
	"flag"

	v1 "github.com/ryanreadbooks/whimer/counter/api/v1"
	"github.com/ryanreadbooks/whimer/counter/internal/config"
	"github.com/ryanreadbooks/whimer/counter/internal/job"
	"github.com/ryanreadbooks/whimer/counter/internal/rpc"
	"github.com/ryanreadbooks/whimer/counter/internal/svc"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor/checker"
	"google.golang.org/grpc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "etc/counter.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx := svc.NewServiceContext(&c)

	server := zrpc.MustNewServer(c.Grpc, func(s *grpc.Server) {
		v1.RegisterCounterServiceServer(s, rpc.NewCounterServer(ctx))
		xgrpc.EnableReflectionIfNecessary(c.Grpc, s)
	})
	interceptor.InstallUnaryServerInterceptors(server,
		interceptor.WithUnaryChecker(checker.UidExistence))

	syncer := job.MustNewSyncer(&c, ctx)

	logx.Infof("counter is serving on %s", c.Grpc.ListenOn)
	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(server)
	group.Add(syncer)
	group.Start()
}
