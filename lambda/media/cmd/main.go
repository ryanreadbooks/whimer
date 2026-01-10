package main

import (
	"context"
	"flag"

	"github.com/ryanreadbooks/whimer/lambda/media/internal/config"
	"github.com/ryanreadbooks/whimer/lambda/media/internal/worker"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	zeroservice "github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/media.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())
	config.MustInit()
	logx.MustSetup(config.Conf.Log)
	defer logx.Close()

	w, err := worker.NewWorker(&config.Conf)
	if err != nil {
		panic(err)
	}
	group := zeroservice.NewServiceGroup()
	defer group.Stop()

	group.Add(workerService{w: w})

	logx.Info("lambda-media worker is serving...")
	group.Start()
}

type workerService struct {
	w *worker.Worker
}

func (s workerService) Start() {
	s.w.Run(context.Background()) // block here
}

func (s workerService) Stop() {
	s.w.Stop()
}
