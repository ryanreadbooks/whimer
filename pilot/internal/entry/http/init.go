package http

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	innerhandler "github.com/ryanreadbooks/whimer/pilot/internal/entry/http/inner/handler"
	innerrouter "github.com/ryanreadbooks/whimer/pilot/internal/entry/http/inner/router"
	xhandler "github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/handler"
	xrouter "github.com/ryanreadbooks/whimer/pilot/internal/entry/http/x/router"

	zeroservice "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

func Register(engine *rest.Server, conf *config.Config, bizz *biz.Biz) {
	xrouter.RegisterX(engine, xhandler.NewHandler(conf, bizz))
	innerrouter.RegisterInner(engine, innerhandler.NewHandler(conf, bizz))

	mod := conf.Http.Mode
	if mod == zeroservice.DevMode || mod == zeroservice.TestMode {
		engine.PrintRoutes()
	}
}
