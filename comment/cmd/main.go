package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/comment/internal/job"
	"github.com/ryanreadbooks/whimer/comment/internal/job/cron"
	"github.com/ryanreadbooks/whimer/comment/internal/rpc"
	"github.com/ryanreadbooks/whimer/comment/internal/svc"
	sdk "github.com/ryanreadbooks/whimer/comment/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
)

var configFile = flag.String("f", "etc/comment.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	ctx := svc.NewServiceContext(&c)

	server := zrpc.MustNewServer(c.Grpc, func(s *grpc.Server) {
		sdk.RegisterReplyServer(s, rpc.NewReplyServer(ctx))
		if c.Grpc.Mode == service.DevMode || c.Grpc.Mode == service.TestMode {
			reflection.Register(s)
		}
	})
	server.AddUnaryInterceptors(
		interceptor.ServerErrorHandle,
		interceptor.ServerMetadataExtract,
		interceptor.ServerValidateHandle,
	)

	mq := kq.MustNewQueue(c.Kafka.AsKqConf(), job.New(ctx))
	csyncer := cron.MustNewCacheSyncer(c.Cron.SyncReplySpec, ctx)

	logx.Info("comment is serving...")
	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(csyncer)
	group.Add(server)
	group.Add(mq)
	group.Start()
}
