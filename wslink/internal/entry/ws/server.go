package ws

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xhttp/middleware"
	"github.com/ryanreadbooks/whimer/passport/pkg/middleware/auth"
	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	"github.com/ryanreadbooks/whimer/wslink/internal/global"
	"github.com/ryanreadbooks/whimer/wslink/internal/infra/dep"
	modelws "github.com/ryanreadbooks/whimer/wslink/internal/model/ws"
	"github.com/ryanreadbooks/whimer/wslink/internal/srv"
	"github.com/zeromicro/go-zero/rest"
)

var (
	_ = auth.MustAuther
	_ = dep.Init
)

type Server struct {
	upgrader    *websocket.Upgrader
	conf        *config.Websocket
	serv        *srv.Service
	sessHandler modelws.SessionHandler

	// server state
	startAt  time.Time // 启动时间
	isClosed atomic.Bool
}

func New(c *config.Config, restServer *rest.Server, service *srv.Service) *Server {
	s := &Server{
		conf:        c.WsServer,
		sessHandler: service,
		serv:        service,
	}

	// http
	subGroup := xhttp.NewRouterGroup(restServer)
	subGroup.Get("/web/sub", s.upgrade,
		middleware.Recovery,
		// auth.UserWeb(dep.Auther()),
	)

	s.upgrader = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	return s
}

// 协议升级成websocket
func (s *Server) upgrade(w http.ResponseWriter, r *http.Request) {
	wsConn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		// err != nil时 upgrader.Upgrade已经处理了
		return
	}

	session := modelws.CreateSession(
		wsConn,
		modelws.WithReadTimeout(s.conf.ReadTimeout.Duration()),
		modelws.WithWriteTimeout(s.conf.WriteTimeout.Duration()),
	)
	session.SetDevice(modelws.DeviceWeb)

	ctx := r.Context()
	if err := s.sessHandler.OnCreate(ctx, session); err != nil {
		// err is not nil, we deny this connection
		modelws.RecoverSession(session)
		xhttp.Error(r, w, err)
		wsConn.Close()
		return
	}

	session.SetOnData(s.sessHandler)
	session.SetOnClose(s.sessHandler)

	concurrent.SafeGo(func() {
		ctx, cancel := context.WithCancel(context.Background())
		defer func() {
			cancel()
			modelws.RecoverSession(session)
		}()

		session.Loop(ctx)
	})
}

// 判断请求是否能升级
func (s *Server) isUpgradeAllowed(r *http.Request) error {
	if r == nil {
		return global.ErrBizInternal
	}

	// 判断最大连接情况

	return nil
}

func (s *Server) Start() {
}

func (s *Server) Stop() {
}
