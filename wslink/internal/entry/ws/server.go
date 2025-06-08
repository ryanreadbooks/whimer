package ws

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xhttp/middleware"
	"github.com/ryanreadbooks/whimer/passport/pkg/middleware/auth"
	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	"github.com/ryanreadbooks/whimer/wslink/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/wslink/internal/srv"
	"github.com/zeromicro/go-zero/rest"
)

type Server struct {
	upgrader *websocket.Upgrader
	conf     *config.Websocket
	sessServ *srv.SessionService

	// server state
	startAt  time.Time // 启动时间
	isClosed atomic.Bool
}

func New(c *config.Config, restServer *rest.Server, serv *srv.Service) *Server {
	s := &Server{
		conf:     c.WsServer,
		sessServ: serv.SessionService,
	}

	// http
	subGroup := xhttp.NewRouterGroup(restServer)
	subGroup.Get("/web/sub", s.upgrade,
		middleware.Recovery,
		auth.UserWeb(dep.Auther()),
	)

	s.upgrader = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	return s
}

func (s *Server) Start() {}

func (s *Server) Stop() {
	ctx, cancel := context.WithTimeout(
		context.Background(), time.Second*time.Duration(config.Conf.System.Shutdown.WaitTime))
	defer cancel()

	s.sessServ.Close(ctx)
}
