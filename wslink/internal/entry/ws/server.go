package ws

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/wslink/internal/config"
	"github.com/ryanreadbooks/whimer/wslink/internal/global"
)

type Server struct {
	httpServer *http.Server
	upgrader   *websocket.Upgrader
	conf       *config.Websocket
	engine     *gin.Engine

	// server state
	startAt  time.Time // 启动时间
	closed   chan struct{}
	isClosed atomic.Bool
}

func New(c *config.Websocket) *Server {
	s := &Server{
		conf: c,
	}

	// http
	s.engine = gin.New()
	s.engine.GET("/sub", s.upgrade)
	// mux := http.NewServeMux()
	// mux.HandleFunc("/sub", s.upgrade)
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.conf.Addr, s.conf.Port),
		Handler: s.engine,
	}
	s.upgrader = &websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	s.closed = make(chan struct{}, 1)

	return s
}

// 协议升级成websocket
func (s *Server) upgrade(c *gin.Context) {
	defer func() {
		if e := recover(); e != nil {
			// 升级过程中panic
		}
	}()

	gwsc, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		// err != nil时 upgrader.Upgrade已经处理了
		return
	}

	wsConnId := uuid.NewString()
	wsConn := GetWsConn(wsConnId, gwsc)
	wsConn.SetReadTimeout(time.Duration(s.conf.ReadTimeout.Duration()))
	wsConn.SetWriteTimeout(time.Duration(s.conf.WriteTimeout.Duration()))
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
	s.startAt = time.Now()
	err := s.httpServer.ListenAndServe()
	if err != nil {
		xlog.Error(fmt.Sprintf("listen and serve failed: %v", err)).Do()
	}
}

func (s *Server) Stop() {
	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		xlog.Error(fmt.Sprintf("close failed: %v", err)).Do()
	}
}
