package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/counter/internal/config"
	grpc "github.com/ryanreadbooks/whimer/counter/internal/entry/grpc"
	"github.com/ryanreadbooks/whimer/counter/internal/infra"
	"github.com/ryanreadbooks/whimer/counter/internal/job"
	"github.com/ryanreadbooks/whimer/counter/internal/srv"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/counter.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())
	infra.Init(&config.Conf)
	svc := srv.NewService(&config.Conf)
	server := grpc.Init(config.Conf.Grpc, svc)

	syncer := job.MustNewSyncer(&config.Conf, svc)
	logx.Infof("counter is serving on %s", config.Conf.Grpc.ListenOn)
	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(server)
	group.Add(syncer)
	group.Start()
}
