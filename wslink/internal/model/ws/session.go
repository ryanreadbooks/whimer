package ws

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/ryanreadbooks/whimer/misc/stacktrace"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

var (
	ErrSessionClosed = fmt.Errorf("session is already closed")
	ErrFinishSession = fmt.Errorf("finish session")
)

func isTimeoutErr(err error) bool {
	if netError, ok := err.(net.Error); ok && netError.Timeout() {
		return true
	}

	return false
}

type Device string

const (
	DeviceWeb Device = "web"
)

// 对connection的封装
type Session struct {
	conn   *connection
	closed atomic.Bool
	device Device

	// callback handler
	onData  SessionOnDataHandler
	onClose SessionOnClosedHandler
}

func (s *Session) reset() {
	s.conn = nil
	s.onClose = nil
	s.onData = nil
	s.closed.Store(true)
	s.device = ""
}

var sessionPool = sync.Pool{
	New: func() any {
		return new(Session)
	},
}

type sessionOpt struct {
	readTimeout  time.Duration
	writeTimeout time.Duration
}

type createSessionOpt func(o *sessionOpt)

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
func CreateSession(webconn *websocket.Conn, opts ...createSessionOpt) *Session {
	o := &sessionOpt{
		readTimeout:  time.Second * 60,
		writeTimeout: time.Second * 60,
	}
	for _, opt := range opts {
		opt(o)
	}

	cid := uuid.NewString()
	c := getWsConn(cid, webconn)
	c.rTimeout = o.readTimeout
	c.wTimeout = o.writeTimeout

	s := sessionPool.Get().(*Session)
	s.conn = c
	s.closed.Store(false)

	return s
}

// 回收session
func RecoverSession(s *Session) {
	if s != nil {
		putWsConn(s.conn)
		s.reset()
		sessionPool.Put(s)
	}
}

func (s *Session) GetId() string {
	return s.conn.id
}

// 消息处理循环
func (s *Session) Loop(ctx context.Context) {
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
		data, err := s.conn.read()
		if err != nil {
			if !errors.Is(err, ErrContinued) {
				// unexpected error should close session
				if isTimeoutErr(err) {
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
			xlog.Msg(fmt.Sprintf("conn %s write error", s.conn.id)).
				Err(err).
				Errorx(ctx)
		}
	}
}

// 处理上行数据
func (s *Session) handleData(ctx context.Context, data []byte) error {
	if s.onData != nil {
		return s.onData.OnData(ctx, s, data)
	}

	return nil
}

func (s *Session) Close(ctx context.Context) {
	cid := s.conn.id
	if s.conn != nil {
		if err := s.conn.close(); err != nil {
			xlog.Msgf("conn %s close err", s.conn.id).Err(err).Errorx(ctx)
		}
	}

	if s.onClose != nil {
		s.onClose.OnClosed(ctx, cid)
	}

	s.closed.Store(true)
}

func (s *Session) GraceClose(ctx context.Context) {
	cid := s.conn.id
	if s.conn != nil {
		if err := s.conn.graceClose(); err != nil {
			xlog.Msgf("conn %s grace close err", s.conn.id).Err(err).Errorx(ctx)
		}
	}

	if s.onClose != nil {
		s.onClose.OnClosed(ctx, cid)
	}

	s.closed.Store(true)
}

func (s *Session) Write(data []byte) error {
	if s.closed.Load() {
		return ErrSessionClosed
	}
	return s.conn.write(data)
}

func (s *Session) WriteText(text string) error {
	if s.closed.Load() {
		return ErrSessionClosed
	}

	return s.conn.writeText(text)
}

func (s *Session) SetOnData(h SessionOnDataHandler) {
	s.onData = h
}

func (s *Session) SetOnClose(h SessionOnClosedHandler) {
	s.onClose = h
}

func (s *Session) SetDevice(dev Device) {
	s.device = dev
}
