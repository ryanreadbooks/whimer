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

	// 获取 data 层实例
	dt := infra.Data()

	// 创建 biz 层，注入 data 依赖
	bizz := biz.New(dt)
	svc := srv.MustNewService(&config.Conf, bizz, dt)

	var err error
	switch *handleType {
	case NoteSyncToEs:
		err = notesynces.Handle(&config.Conf, bizz, svc, dt)
	default:
		xlog.Msgf("unsupported handle type: %s", *handleType).Error()
		os.Exit(1)
	}

	if err != nil {
		xlog.Msgf("notesynces failed").Err(err).Error()
		os.Exit(1)
	}
}
