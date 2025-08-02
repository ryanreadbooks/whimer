package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	backend "github.com/ryanreadbooks/whimer/api-x/internal/handler"
	"github.com/ryanreadbooks/whimer/api-x/internal/router"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	zeroservice "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/api-x.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	backend.Init(&c)
	var handler = backend.NewHandler(&c)

	apiserver := rest.MustNewServer(c.Http)
	router.RegX(apiserver, handler)

	servgroup := zeroservice.NewServiceGroup()
	defer servgroup.Stop()
	servgroup.Add(apiserver)

	logx.Info("api-x server is running...")
	servgroup.Start()
}
