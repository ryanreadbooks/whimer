package main

import (
	"flag"
	"fmt"

	"github.com/ryanreadbooks/whimer/passport/internal/config"
	"github.com/ryanreadbooks/whimer/passport/internal/entry/grpc"
	"github.com/ryanreadbooks/whimer/passport/internal/entry/http"
	"github.com/ryanreadbooks/whimer/passport/internal/srv"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/passport.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())
	if err := config.Conf.Init(); err != nil {
		panic(fmt.Errorf("panic: config init: %w", err))
	}

	s := srv.New(&config.Conf)
	restServer := rest.MustNewServer(config.Conf.Http)
	http.Init(restServer, s)

	grpcServer := grpc.Init(config.Conf.Grpc, s)

	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(restServer)
	group.Add(grpcServer)

	logx.Info("passport is serving...")
	group.Start()
}
