package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/misc/must"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/http/handler"
	httprouter "github.com/ryanreadbooks/whimer/pilot/internal/entry/http/router"
	"github.com/ryanreadbooks/whimer/pilot/internal/entry/messaging"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra"
	"github.com/ryanreadbooks/whimer/pilot/internal/job"
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

	bizz := biz.New(&config.Conf)
	messaging.Init(&config.Conf, bizz)
	defer messaging.Close()

	var handler = handler.NewHandler(&config.Conf, bizz)
	apiserver := rest.MustNewServer(config.Conf.Http)
	httprouter.RegisterX(apiserver, handler)

	servgroup := zeroservice.NewServiceGroup()
	defer servgroup.Stop()

	noteEvtJob := job.NewNoteEventManager(
		job.NoteEventManagerConfig{
			Tick: config.Conf.JobConfig.NoteEventJob.Interval,
		},
		bizz,
	)

	servgroup.Add(apiserver)
	servgroup.Add(noteEvtJob)

	logx.Info("pilot server is running...")
	servgroup.Start()
}
