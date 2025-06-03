package ws

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// 通用错误
var (
	ErrConnClosed         = fmt.Errorf("conn closed")
	ErrUnsupportedMsgType = fmt.Errorf("unsupported msg type")
	ErrContinued          = fmt.Errorf("conn continued")
	ErrUseOfClosedConn    = fmt.Errorf("use of closed network connection")
)

var wsConnPool = sync.Pool{
	New: func() any {
		return new(WsConn)
	},
}

func GetWsConn(id string, wc *websocket.Conn) *WsConn {
	c, _ := wsConnPool.Get().(*WsConn)
	c.id = id
	c.conn = wc
	return c
}

func PutWsConn(c *WsConn) {
	if c != nil {
		c.Reset()
		wsConnPool.Put(c)
	}
}

// websocket连接
type WsConn struct {
	id   string
	conn *websocket.Conn

	rTimeout time.Duration
	wTimeout time.Duration
}

func (c *WsConn) Reset() {
	c.id = ""
	c.conn = nil
	c.rTimeout = 0
	c.wTimeout = 0
}

func (c *WsConn) Read() ([]byte, error) {
	msgTyp, data, err := c.conn.ReadMessage()
	if err != nil {
		if strings.Contains(err.Error(), net.ErrClosed.Error()) {
			// use of closed network connection
			return nil, ErrUseOfClosedConn
		}
		return nil, err
	}

	if msgTyp == websocket.TextMessage {
		return nil, ErrUnsupportedMsgType
	}

	if msgTyp == websocket.PingMessage {
		// pong back
		err = c.conn.WriteMessage(websocket.PongMessage, nil)
		if err != nil {
			return nil, ErrContinued
		}
	}

	if msgTyp == websocket.CloseMessage {
		return nil, ErrConnClosed
	}

	// binary data
	return data, err
}

func (c *WsConn) Write(data []byte) error {
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *WsConn) Close() error {
	return c.conn.Close()
}

func (c *WsConn) CloseWhenServerErr(msg string) error {
	errW := c.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseInternalServerErr, msg))
	errC := c.conn.Close()
	return errors.Join(errW, errC)
}

func (c *WsConn) SetConn(conn *websocket.Conn) {
	c.conn = conn
}

func (c *WsConn) SetReadTimeout(timeout time.Duration) error {
	return c.conn.SetReadDeadline(time.Now().Add(timeout))
}

func (c *WsConn) SetWriteTimeout(timeout time.Duration) error {
	return c.conn.SetWriteDeadline(time.Now().Add(timeout))
}

func (c *WsConn) GetId() string {
	return c.id
}

func (c *WsConn) GetRemote() string {
	return c.conn.RemoteAddr().String()
}
