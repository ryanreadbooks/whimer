package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/relation/internal/config"
	"github.com/ryanreadbooks/whimer/relation/internal/entry/grpc"
	"github.com/ryanreadbooks/whimer/relation/internal/infra"
	"github.com/ryanreadbooks/whimer/relation/internal/srv"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/relation.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())
	logx.MustSetup(config.Conf.Log)
	defer logx.Close()

	// 基础设施初始化
	infra.Init(&config.Conf)
	defer infra.Close()

	svc := srv.NewService(&config.Conf)

	grpcServer := grpc.Init(config.Conf.Grpc, svc)

	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(grpcServer)
	logx.Info("relation is serving...")
	group.Start()
}
