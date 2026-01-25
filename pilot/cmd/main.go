package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/misc/must"
	"github.com/ryanreadbooks/whimer/pilot/internal/app"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/messaging"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	zeroservice "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/pilot.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())
	must.Do(config.Conf.Init())
	logx.MustSetup(config.Conf.Log)
	defer logx.Close()

	infra.Init(&config.Conf)
	defer infra.Close()

	appMgr := app.NewManager(&config.Conf)

	bizz := biz.New(&config.Conf)
	messaging.Init(&config.Conf, bizz, appMgr)
	defer messaging.Close()

	apiserver := rest.MustNewServer(config.Conf.Http)
	http.Register(apiserver, &config.Conf, bizz, appMgr)

	servgroup := zeroservice.NewServiceGroup()
	defer servgroup.Stop()

	servgroup.Add(apiserver)

	logx.Info("pilot server is running...")
	servgroup.Start()
}
