package http

import (
	"github.com/ryanreadbooks/whimer/note/internal/config"
	innerhandler "github.com/ryanreadbooks/whimer/note/internal/entry/http/inner/handler"
	innerrouter "github.com/ryanreadbooks/whimer/note/internal/entry/http/inner/router"
	"github.com/ryanreadbooks/whimer/note/internal/srv"

	zeroservice "github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

func Init(c rest.RestConf, svc *srv.Service) *rest.Server {
	engine := rest.MustNewServer(c)
	Register(engine, svc)
	return engine
}

func Register(engine *rest.Server, svc *srv.Service) {
	innerrouter.RegisterInner(engine, innerhandler.NewHandler(svc))

	mod := config.Conf.Http.Mode
	if mod == zeroservice.DevMode || mod == zeroservice.TestMode {
		engine.PrintRoutes()
	}
}
