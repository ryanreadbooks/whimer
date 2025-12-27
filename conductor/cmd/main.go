package main

import (
	"context"
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
	infra.Init(&config.Conf)

	rootCtx, rootCancel := context.WithCancel(context.Background())
	bizz := biz.NewBiz(rootCtx, &config.Conf)
	srv := service.NewService(&config.Conf, bizz)
	bizz.Start()
	srv.Start(rootCtx)
	defer func() {
		rootCancel()

		srv.Stop()
		bizz.Stop()
		infra.Close()
		logx.Close()
	}()

	server := grpc.Init(config.Conf.Grpc, srv)
	group := zeroservice.NewServiceGroup()
	defer group.Stop()
	// added service的启动顺序不能保证 所以biz和service单独处理
	group.Add(server)

	logx.Info("conductor is serving...")
	group.Start()
}
