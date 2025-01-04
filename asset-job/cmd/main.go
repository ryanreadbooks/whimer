package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/asset-job/internal/config"
	"github.com/ryanreadbooks/whimer/asset-job/internal/entry/events"
	"github.com/ryanreadbooks/whimer/asset-job/internal/srv"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	gzservice "github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/asset-job.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())

	svc := srv.NewService(&config.Conf)

	logx.Info("asset-job is running...")
	group := gzservice.NewServiceGroup()
	defer group.Stop()

	eventQs := events.Init(&config.Conf, svc)

	for _, q := range eventQs {
		group.Add(q)
	}

	group.Start()
}
