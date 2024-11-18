package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/feed/internal/config"
	"github.com/ryanreadbooks/whimer/feed/internal/entry/http"

	"github.com/zeromicro/go-zero/core/conf"
	zeroservice "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/core/logx"
)

var configFile = flag.String("f", "etc/feed.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())

	apiserver := rest.MustNewServer(config.Conf.Http)
	http.Init(apiserver)

	servgroup := zeroservice.NewServiceGroup()
	defer servgroup.Stop()
	servgroup.Add(apiserver)

	logx.Info("feed http server is running...")
	servgroup.Start()
}
