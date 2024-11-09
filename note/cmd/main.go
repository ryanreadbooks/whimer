package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/entry/grpc"
	"github.com/ryanreadbooks/whimer/note/internal/srv"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/note.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())
	srv := srv.NewService(&config.Conf)

	grpcServer := grpc.Init(config.Conf.Grpc, srv)

	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(grpcServer)
	logx.Info("note is serving...")
	group.Start()
}
