package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/search/internal/config"
	"github.com/ryanreadbooks/whimer/search/internal/entry/grpc"
	"github.com/ryanreadbooks/whimer/search/internal/entry/messaging"
	"github.com/ryanreadbooks/whimer/search/internal/infra"
	"github.com/ryanreadbooks/whimer/search/internal/srv"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/search.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())

	infra.Init(&config.Conf)
	defer infra.Close()
	svc := srv.NewService(&config.Conf)
	defer svc.Stop()

	messaging.Init(&config.Conf, svc)
	defer messaging.Close()

	grpcServer := grpc.Init(config.Conf.Grpc, svc)
	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(grpcServer)
	logx.Info("search is serving...")
	group.Start()
}
