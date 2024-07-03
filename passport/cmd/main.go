package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/passport/internal/config"
	"github.com/ryanreadbooks/whimer/passport/internal/handler"
	"github.com/ryanreadbooks/whimer/passport/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/passport.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	ctx := svc.NewServiceContext(&c)
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	handler.RegisterHandlers(server, ctx)
	logx.Info("passport is serving...")
	server.Start()
}
