package ws

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ryanreadbooks/whimer/misc/stacktrace"
	"github.com/ryanreadbooks/whimer/misc/utils"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/wslink/internal/model"
)

var (
	ErrSessionClosed = fmt.Errorf("connection is already closed")
	ErrFinishSession = fmt.Errorf("finish session")

	// websocket通用错误
	ErrWebsocketConnClosed = fmt.Errorf("websocket conn closed")
	ErrUnsupportedMsgType  = fmt.Errorf("unsupported msg type")
	ErrContinued           = fmt.Errorf("conn continued")
	ErrUseOfClosedConn     = fmt.Errorf("use of closed network connection")
)

func isTimeoutErr(err error) bool {
	if netError, ok := err.(net.Error); ok && netError.Timeout() {
		return true
	}

	return false
}

// 对connection的封装
type Connection struct {
	id     string
	conn   *websocket.Conn
	closed atomic.Bool
	device model.Device

	rTimeout time.Duration
	wTimeout time.Duration

	// callback handler
	onData      ConnectionOnDataHandler
	afterClosed ConnectionAfterClosedHandler
}

func (s *Connection) reset() {
	s.id = ""
	s.conn = nil
	s.closed.Store(true)
	s.device = ""

	s.rTimeout = 0
	s.wTimeout = 0

	s.afterClosed = nil
	s.onData = nil
}

var sessionPool = sync.Pool{
	New: func() any {
		return new(Connection)
	},
}

type sessionOpt struct {
	autoId       bool
	readTimeout  time.Duration
	writeTimeout time.Duration
}

type createSessionOpt func(o *sessionOpt)

func WithAutoId(b bool) createSessionOpt {
	return func(o *sessionOpt) {
		o.autoId = b
	}
}

func WithReadTimeout(rt time.Duration) createSessionOpt {
	return func(o *sessionOpt) {
		o.readTimeout = rt
	}
}

func WithWriteTimeout(wt time.Duration) createSessionOpt {
	return func(o *sessionOpt) {
		o.writeTimeout = wt
	}
}

// 获取一个session实例
func CreateSession(webconn *websocket.Conn, opts ...createSessionOpt) *Connection {
	o := &sessionOpt{
		autoId:       true,
		readTimeout:  time.Second * 60,
		writeTimeout: time.Second * 60,
	}
	for _, opt := range opts {
		opt(o)
	}

	s := sessionPool.Get().(*Connection)
	s.conn = webconn
	s.rTimeout = o.readTimeout
	s.wTimeout = o.writeTimeout
	if o.autoId {
		cid := uuid.NewString()
		s.id = cid
	}

	s.closed.Store(false)

	return s
}

// 回收session
func RecoverSession(s *Connection) {
	if s != nil {
		s.reset()
		sessionPool.Put(s)
	}
}

func (s *Connection) GetId() string {
	return s.id
}

func (s *Connection) SetId(id string) {
	s.id = id
}

func (s *Connection) read() ([]byte, error) {
	s.conn.SetReadDeadline(time.Now().Add(s.rTimeout))
	msgTyp, data, err := s.conn.ReadMessage()
	if err != nil {
		if strings.Contains(err.Error(), net.ErrClosed.Error()) {
			// use of closed network connection
			return nil, ErrUseOfClosedConn
		}
		return nil, err
	}

	if msgTyp == websocket.PingMessage {
		// pong back
		err = s.conn.WriteMessage(websocket.PongMessage, []byte("PONG"))
		if err != nil {
			return nil, ErrContinued
		}
	}

	if msgTyp == websocket.PongMessage {
		return nil, ErrContinued
	}

	if msgTyp == websocket.CloseMessage {
		return nil, ErrWebsocketConnClosed
	}

	// binary data
	return data, err
}

// 消息处理循环
func (s *Connection) Loop(ctx context.Context) {
	defer func() {
		if e := recover(); e != nil {
			logErr := xerror.Wrapf(xerror.ErrPanic, fmt.Sprintf("%v", e))
			xlog.Msg("panic").
				Err(logErr).
				Extra("stack", stacktrace.FormatFrames(xerror.UnwrapFrames(logErr))).
				Errorx(ctx)
			s.Close(ctx)
		}
	}()

	for {
		data, err := s.read()
		if err != nil {
			if !errors.Is(err, ErrContinued) {
				// unexpected error should close session
				if isTimeoutErr(err) {
					// After a read has timed out, the websocket connection state is corrupt and
					// all future reads will return an error
					s.GraceClose(ctx)
				} else {
					s.Close(ctx)
				}
				return
			}

			continue
		}

		err = s.handleData(ctx, data)
		if err != nil {
			if errors.Is(err, ErrFinishSession) {
				// ErrFinishSession will close this session
				s.Close(ctx)
				return
			}
			// 其它情况记录日志
			xlog.Msg(fmt.Sprintf("conn %s write error", s.id)).
				Err(err).
				Errorx(ctx)
		}
	}
}

// 处理上行数据
func (s *Connection) handleData(ctx context.Context, data []byte) error {
	if s.onData != nil {
		return s.onData.OnData(ctx, s, data)
	}

	return nil
}

func (s *Connection) Close(ctx context.Context) {
	cid := s.id
	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			xlog.Msgf("conn %s close err", s.id).Err(err).Errorx(ctx)
		}
	}

	if s.afterClosed != nil {
		s.afterClosed.AfterClosed(ctx, cid)
	}

	s.closed.Store(true)
}

func (s *Connection) GraceClose(ctx context.Context) {
	cid := s.id
	if s.conn != nil {
		err1 := s.conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "CLOSED"),
			time.Time{},
		)
		err2 := s.conn.Close()
		err := errors.Join(err1, err2)
		if err != nil {
			xlog.Msgf("conn %s grace close err", s.id).Err(err).Errorx(ctx)
		}
	}

	if s.afterClosed != nil {
		s.afterClosed.AfterClosed(ctx, cid)
	}

	s.closed.Store(true)
}

func (s *Connection) Write(data []byte) error {
	if s.closed.Load() {
		return ErrSessionClosed
	}

	s.conn.SetWriteDeadline(time.Now().Add(s.wTimeout))
	return s.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (s *Connection) WriteText(text string) error {
	if s.closed.Load() {
		return ErrSessionClosed
	}

	s.conn.SetWriteDeadline(time.Now().Add(s.wTimeout))
	return s.conn.WriteMessage(websocket.TextMessage, utils.StringToBytes(text))
}

func (s *Connection) SetOnData(h ConnectionOnDataHandler) {
	s.onData = h
}

func (s *Connection) SetAfterClosed(h ConnectionAfterClosedHandler) {
	s.afterClosed = h
}

func (s *Connection) SetDevice(dev model.Device) {
	s.device = dev
}

func (s *Connection) GetDevice() model.Device {
	return s.device
}

func (s *Connection) GetRemote() string {
	return s.conn.RemoteAddr().String()
}
