package main

import (
	"flag"
	"fmt"

	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/entry/grpc"
	"github.com/ryanreadbooks/whimer/note/internal/entry/http"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	"github.com/ryanreadbooks/whimer/note/internal/srv"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/note.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())
	logx.MustSetup(config.Conf.Log)
	defer logx.Close()

	if err := config.Conf.Init(); err != nil {
		panic(fmt.Errorf("panic: config init: %w", err))
	}
	global.MustInit(&config.Conf)
	infra.Init(&config.Conf)
	defer infra.Close()

	// 获取 data 层实例
	dt := infra.Data()

	// 创建 biz 层，注入 data 依赖
	bizz := biz.New(dt)

	// 创建 service 层
	svc := srv.MustNewService(&config.Conf, bizz, dt)

	grpcServer := grpc.Init(config.Conf.Grpc, svc)
	httpServer := http.Init(config.Conf.Http, svc)

	group := service.NewServiceGroup()
	defer group.Stop()

	group.Add(svc)
	group.Add(grpcServer)
	group.Add(httpServer)
	logx.Info("note is serving...")
	group.Start()
}
