package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/conductor/internal/biz"
	"github.com/ryanreadbooks/whimer/conductor/internal/config"
	"github.com/ryanreadbooks/whimer/conductor/internal/entry/grpc"
	"github.com/ryanreadbooks/whimer/conductor/internal/global"
	"github.com/ryanreadbooks/whimer/conductor/internal/infra"
	"github.com/ryanreadbooks/whimer/conductor/internal/service"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	zeroservice "github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/conductor.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())
	global.MustInit(&config.Conf)
	logx.MustSetup(config.Conf.Log)
	defer logx.Close()

	infra.Init(&config.Conf)
	defer infra.Close()

	bizz := biz.NewBiz(&config.Conf)

	srv := service.NewService(&config.Conf, bizz)
	server := grpc.Init(config.Conf.Grpc, srv)
	group := zeroservice.NewServiceGroup()
	defer group.Stop()

	group.Add(bizz)
	group.Add(server)

	logx.Info("conductor is serving...")
	group.Start()
}
