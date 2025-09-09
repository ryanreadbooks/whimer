package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/note/cmd/handle/notesynces"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	"github.com/ryanreadbooks/whimer/note/internal/srv"
	"github.com/zeromicro/go-zero/core/conf"
)

const (
	NoteSyncToEs = "notesynces"
)

var (
	configFile = flag.String("f", "etc/note.yaml", "the config file")
	handleType = flag.String("handle-type", NoteSyncToEs, "job type")
)

func main() {
	flag.Parse()

	conf.MustLoad(*configFile, &config.Conf, conf.UseEnv())
	if err := config.Conf.Init(); err != nil {
		panic(fmt.Errorf("panic: config init: %w", err))
	}
	infra.Init(&config.Conf)
	defer infra.Close()

	bizz := biz.New()
	svc := srv.NewService(&config.Conf, bizz)

	var err error
	switch *handleType {
	case NoteSyncToEs:
		err = notesynces.Handle(&config.Conf, bizz, svc)
	default:
		xlog.Msgf("unsupported handle type: %s", *handleType).Error()
		os.Exit(1)
	}

	if err != nil {
		xlog.Msgf("notesynces failed").Err(err).Error()
		os.Exit(1)
	}
}
