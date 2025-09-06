package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	httpbackend "github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler"
	httprouter "github.com/ryanreadbooks/whimer/api-x/internal/entry/http/router"
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
	httpbackend.Init(&c)
	var handler = httpbackend.NewHandler(&c)

	apiserver := rest.MustNewServer(c.Http)
	httprouter.RegisterX(apiserver, handler)

	servgroup := zeroservice.NewServiceGroup()
	defer servgroup.Stop()
	servgroup.Add(apiserver)

	logx.Info("api-x server is running...")
	servgroup.Start()
}
