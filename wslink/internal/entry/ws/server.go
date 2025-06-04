package ws

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xhttp/middleware"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	"github.com/ryanreadbooks/whimer/wslink/internal/global"
	modelws "github.com/ryanreadbooks/whimer/wslink/internal/model/ws"
	"github.com/ryanreadbooks/whimer/wslink/internal/srv"
	"github.com/zeromicro/go-zero/rest"
)

type Server struct {
	httpServer  *http.Server
	upgrader    *websocket.Upgrader
	conf        *config.Websocket
	engine      *gin.Engine
	sessHandler modelws.SessionHandler

	// server state
	startAt  time.Time // 启动时间
	closed   chan struct{}
	isClosed atomic.Bool
}

func New(c *config.Config, restServer *rest.Server, service *srv.Service) *Server {
	s := &Server{
		conf:        c.WsServer,
		sessHandler: service,
	}

	// http
	subGroup := xhttp.NewRouterGroup(restServer)
	subGroup.Get("/sub", s.upgrade, middleware.Recovery)
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", c.Http.Host, c.Http.Port),
		Handler: s.engine,
	}
	s.upgrader = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	s.closed = make(chan struct{}, 1)

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
	// TODO close all existing sessions
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		xlog.Error(fmt.Sprintf("close failed: %v", err)).Do()
	}
}
