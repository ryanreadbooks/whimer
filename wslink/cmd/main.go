package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	"github.com/ryanreadbooks/whimer/wslink/internal/entry/ws"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/wslink.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	handler := ws.New(c.WsServer)

	group := service.NewServiceGroup()
	group.Add(handler)
	defer group.Stop()
	group.Start()
}
