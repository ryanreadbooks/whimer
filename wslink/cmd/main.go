package main

import (
	"flag"
	"time"

	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	"github.com/ryanreadbooks/whimer/wslink/internal/entry/grpc"
	"github.com/ryanreadbooks/whimer/wslink/internal/entry/ws"
	"github.com/ryanreadbooks/whimer/wslink/internal/infra"
	"github.com/ryanreadbooks/whimer/wslink/internal/srv"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/wslink.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())
	config.Init()
	logx.MustSetup(config.Conf.Log)
	defer logx.Close()

	infra.Init(&config.Conf)
	serv := srv.New(&config.Conf)

	apiServer := rest.MustNewServer(config.Conf.Http)
	wsServer := ws.New(&config.Conf, apiServer, serv)
	grpcServer := grpc.Init(config.Conf.Grpc, serv)

	proc.SetTimeToForceQuit(time.Duration(config.Conf.System.Shutdown.WaitTime) * time.Second)

	group := service.NewServiceGroup()
	group.Add(apiServer)
	group.Add(wsServer)
	group.Add(grpcServer)
	defer group.Stop()

	logx.Info("wslink server is running...")
	group.Start()
}
