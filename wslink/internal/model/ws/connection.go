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
	ErrConnectionClosed = fmt.Errorf("connection is already closed")
	ErrFinishConnection = fmt.Errorf("finish session")
	ErrContinued        = fmt.Errorf("conn continued")

	// websocket通用错误
	ErrWebsocketConnClosed = fmt.Errorf("websocket conn closed")
	ErrUnsupportedMsgType  = fmt.Errorf("unsupported msg type")
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
	id      string
	reqId   string // 来自客户端
	up      int64  // 上行次数记录
	conn    *websocket.Conn
	closed  atomic.Bool
	device  model.Device
	localIp string

	rTimeout time.Duration
	wTimeout time.Duration

	// callback handler
	onData      ConnectionOnDataHandler
	afterClosed ConnectionAfterClosedHandler
}

func (c *Connection) reset() {
	c.id = ""
	c.reqId = ""
	c.conn = nil
	c.closed.Store(true)
	c.device = ""
	c.localIp = ""

	c.rTimeout = 0
	c.wTimeout = 0

	c.afterClosed = nil
	c.onData = nil
}

var connectionPool = sync.Pool{
	New: func() any {
		return new(Connection)
	},
}

type connectionOpt struct {
	autoId       bool
	readTimeout  time.Duration
	writeTimeout time.Duration
}

type createConnectionOpt func(o *connectionOpt)

func WithAutoId(b bool) createConnectionOpt {
	return func(o *connectionOpt) {
		o.autoId = b
	}
}

func WithReadTimeout(rt time.Duration) createConnectionOpt {
	return func(o *connectionOpt) {
		o.readTimeout = rt
	}
}

func WithWriteTimeout(wt time.Duration) createConnectionOpt {
	return func(o *connectionOpt) {
		o.writeTimeout = wt
	}
}

// 获取一个连接实例
func CreateConnection(webconn *websocket.Conn, opts ...createConnectionOpt) *Connection {
	o := &connectionOpt{
		autoId:       true,
		readTimeout:  time.Second * 60,
		writeTimeout: time.Second * 60,
	}
	for _, opt := range opts {
		opt(o)
	}

	s := connectionPool.Get().(*Connection)
	s.conn = webconn
	s.rTimeout = o.readTimeout
	s.wTimeout = o.writeTimeout
	if o.autoId {
		s.id = uuid.NewString()
	}

	s.closed.Store(false)

	return s
}

// 回收
func RecoverConnection(s *Connection) {
	if s != nil {
		s.reset()
		connectionPool.Put(s)
	}
}

func (c *Connection) GetId() string {
	return c.id
}

func (c *Connection) SetId(id string) {
	c.id = id
}

func (c *Connection) read() ([]byte, error) {
	c.conn.SetReadDeadline(time.Now().Add(c.rTimeout))
	msgTyp, data, err := c.conn.ReadMessage()
	if err != nil {
		if strings.Contains(err.Error(), net.ErrClosed.Error()) {
			// use of closed network connection
			return nil, ErrUseOfClosedConn
		}
		return nil, err
	}

	if msgTyp == websocket.PingMessage {
		// pong back
		err = c.conn.WriteMessage(websocket.PongMessage, []byte("PONG"))
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
func (c *Connection) Loop(ctx context.Context) {
	c.up++
	defer func() {
		if e := recover(); e != nil {
			logErr := xerror.Wrapf(xerror.ErrPanic, "%v", e)
			xlog.Msg("panic").
				Err(logErr).
				Extra("stack", stacktrace.FormatFrames(xerror.UnwrapFrames(logErr))).
				Errorx(ctx)
			c.Close(ctx)
		}
	}()

	for !c.closed.Load() {
		data, err := c.read()
		if err != nil {
			if !errors.Is(err, ErrContinued) {
				// unexpected error should close session
				if isTimeoutErr(err) {
					// After a read has timed out, the websocket connection state is corrupt and
					// all future reads will return an error
					c.GraceClose(ctx)
				} else {
					c.Close(ctx)
				}
				return
			}

			continue
		}

		c.up++
		err = c.handleData(ctx, data)
		if err != nil {
			if errors.Is(err, ErrFinishConnection) {
				// ErrFinishSession will close this session
				c.Close(ctx)
				return
			}
			// 其它情况记录日志
			xlog.Msg(fmt.Sprintf("conn %s write error", c.id)).
				Err(err).
				Errorx(ctx)
		}
	}
}

// 处理上行数据
func (c *Connection) handleData(ctx context.Context, data []byte) error {
	if c.onData != nil {
		return c.onData.OnData(ctx, c, data)
	}

	return nil
}

func (c *Connection) Close(ctx context.Context) {
	cid := c.id
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			xlog.Msgf("conn %s close err", c.id).Err(err).Errorx(ctx)
		}
	}

	if c.afterClosed != nil {
		c.afterClosed.AfterClosed(ctx, cid)
	}

	c.closed.Store(true)
}

// GraceClose will send close control through websocket then close the net connection;
func (c *Connection) GraceClose(ctx context.Context) {
	cid := c.id
	if c.conn != nil {
		err1 := c.conn.WriteControl(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "CLOSED"),
			time.Time{},
		)
		err2 := c.conn.Close()
		err := errors.Join(err1, err2)
		if err2 != nil {
			xlog.Msgf("conn %s grace close err", c.id).Err(err).Infox(ctx)
		}
	}

	if c.afterClosed != nil {
		c.afterClosed.AfterClosed(ctx, cid)
	}

	c.closed.Store(true)
}

func (c *Connection) Write(data []byte) error {
	if c.closed.Load() {
		return ErrConnectionClosed
	}

	c.conn.SetWriteDeadline(time.Now().Add(c.wTimeout))
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *Connection) WriteText(text string) error {
	if c.closed.Load() {
		return ErrConnectionClosed
	}

	c.conn.SetWriteDeadline(time.Now().Add(c.wTimeout))
	return c.conn.WriteMessage(websocket.TextMessage, utils.StringToBytes(text))
}

func (c *Connection) SetOnData(h ConnectionOnDataHandler) {
	c.onData = h
}

func (c *Connection) SetAfterClosed(h ConnectionAfterClosedHandler) {
	c.afterClosed = h
}

func (c *Connection) SetDevice(dev model.Device) {
	c.device = dev
}

func (c *Connection) GetDevice() model.Device {
	return c.device
}

func (c *Connection) GetRemote() string {
	return c.conn.RemoteAddr().String()
}

func (c *Connection) GetLocalIp() string {
	return c.localIp
}

func (c *Connection) SetLocalIp(ip string) {
	c.localIp = ip
}

func (c *Connection) SetReqId(reqId string) {
	c.reqId = reqId
}
