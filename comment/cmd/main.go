package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/comment/internal/job"
	cronjob "github.com/ryanreadbooks/whimer/comment/internal/job/cron"
	"github.com/ryanreadbooks/whimer/comment/internal/rpc"
	"github.com/ryanreadbooks/whimer/comment/internal/svc"
	sdk "github.com/ryanreadbooks/whimer/comment/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"google.golang.org/grpc"

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
		sdk.RegisterReplyServiceServer(s, rpc.NewReplyServer(ctx))
		xgrpc.EnableReflection(c.Grpc, s)
	})
	interceptor.InstallServerUnaryInterceptors(server,
		interceptor.WithChecker(interceptor.UidExistenceChecker))

	mq := kq.MustNewQueue(c.Kafka.AsKqConf(), job.New(ctx))
	csyncer := cronjob.MustNewCacheSyncer(c.Cron.SyncReplySpec, ctx)

	logx.Info("comment is serving...")
	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(csyncer)
	group.Add(server)
	group.Add(mq)
	group.Start()
}
