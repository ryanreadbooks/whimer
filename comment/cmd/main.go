package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/comment/internal/entry/grpc"
	"github.com/ryanreadbooks/whimer/comment/internal/srv"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/comment.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())
	svc := srv.NewService(&config.Conf)
	server := grpc.Init(config.Conf.Grpc, svc)

	logx.Info("comment is serving...")
	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(server)
	group.Start()
}
