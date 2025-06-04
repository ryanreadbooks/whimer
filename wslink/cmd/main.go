package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	"github.com/ryanreadbooks/whimer/wslink/internal/entry/ws"
	"github.com/ryanreadbooks/whimer/wslink/internal/srv"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/wslink.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	apiserver := rest.MustNewServer(c.Http)
	serv := srv.NewService(&c)
	handler := ws.New(&c, apiserver, serv)

	group := service.NewServiceGroup()
	group.Add(apiserver)
	group.Add(handler)
	defer group.Stop()

	logx.Info("wslink server is running...")
	group.Start()
}
