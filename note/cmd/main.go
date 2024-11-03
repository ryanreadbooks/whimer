package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/rpc"
	"github.com/ryanreadbooks/whimer/note/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/note.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())
	ctx := svc.NewServiceContext(&config.Conf)

	grpcServer := rpc.Init(config.Conf.Grpc, ctx)

	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(grpcServer)
	logx.Info("note is serving...")
	group.Start()
}
