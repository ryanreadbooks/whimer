package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/api-x/internal/biz"
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/api-x/internal/daemon"
	httpbackend "github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler"
	httprouter "github.com/ryanreadbooks/whimer/api-x/internal/entry/http/router"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	zeroservice "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/api-x.yaml", "the config file")

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())
	httpbackend.Init(&config.Conf)

	bizz := biz.New(&config.Conf)
	var handler = httpbackend.NewHandler(&config.Conf, bizz)

	apiserver := rest.MustNewServer(config.Conf.Http)
	httprouter.RegisterX(apiserver, handler)

	servgroup := zeroservice.NewServiceGroup()
	defer servgroup.Stop()

	noteEvtDaemon := daemon.NewNoteEventManager(
		daemon.NoteEventManagerConfig{
			Tick: config.Conf.DaemonConfig.NoteEventDaemon.Interval,
		}, bizz)

	servgroup.Add(apiserver)
	servgroup.Add(noteEvtDaemon)

	logx.Info("api-x server is running...")
	servgroup.Start()
}
