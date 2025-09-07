package main

import (
	"flag"

	"github.com/ryanreadbooks/whimer/api-x/internal/biz"
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	httpbackend "github.com/ryanreadbooks/whimer/api-x/internal/entry/http/handler"
	httprouter "github.com/ryanreadbooks/whimer/api-x/internal/entry/http/router"
	"github.com/ryanreadbooks/whimer/api-x/internal/job"
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

	noteEvtJob := job.NewNoteEventJobManager(
		job.NoteEventJobManagerConfig{
			Tick: config.Conf.JobConfig.NoteEventJob.Interval,
		}, bizz)

	servgroup.Add(apiserver)
	servgroup.Add(noteEvtJob)

	logx.Info("api-x server is running...")
	servgroup.Start()
}
