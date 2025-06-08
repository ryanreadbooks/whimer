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

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	config.Conf = c
	config.Init()

	infra.Init(&c)
	serv := srv.New(&c)

	apiServer := rest.MustNewServer(c.Http)
	wsServer := ws.New(&c, apiServer, serv)
	grpcServer := grpc.Init(c.Grpc, serv)

	proc.SetTimeToForceQuit(time.Duration(c.System.Shutdown.WaitTime) * time.Second)

	group := service.NewServiceGroup()
	group.Add(apiServer)
	group.Add(wsServer)
	group.Add(grpcServer)
	defer group.Stop()

	logx.Info("wslink server is running...")
	group.Start()
}
