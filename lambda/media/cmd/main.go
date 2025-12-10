package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/lambda/media/internal/config"
	"github.com/ryanreadbooks/whimer/lambda/media/internal/entry/http"
	"github.com/ryanreadbooks/whimer/lambda/media/internal/service"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	zeroservice "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/media.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())
	defer logx.Close()

	srv := service.New(&config.Conf)
	restServer := rest.MustNewServer(config.Conf.Http)
	http.Init(restServer, srv)

	group := zeroservice.NewServiceGroup()
	defer group.Stop()

	group.Add(restServer)

	logx.Info("lambda-media is serving...")
	group.Start()
}
