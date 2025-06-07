package main

import (
	"flag"
	"time"

	"github.com/ryanreadbooks/whimer/wslink/internal/config"
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
	apiserver := rest.MustNewServer(c.Http)
	serv := srv.NewService(&c)
	wshandler := ws.New(&c, apiserver, serv)

	proc.SetTimeToForceQuit(time.Duration(c.System.Shutdown.WaitTime) * time.Second)
	group := service.NewServiceGroup()
	group.Add(apiserver)
	group.Add(wshandler)
	defer group.Stop()

	logx.Info("wslink server is running...")
	group.Start()
}
