package ws

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ryanreadbooks/whimer/misc/utils"
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
		return new(connection)
	},
}

func getWsConn(id string, wc *websocket.Conn) *connection {
	c, _ := wsConnPool.Get().(*connection)
	c.id = id
	c.conn = wc
	return c
}

func putWsConn(c *connection) {
	if c != nil {
		c.reset()
		wsConnPool.Put(c)
	}
}

// websocket连接
type connection struct {
	id   string
	conn *websocket.Conn

	rTimeout time.Duration
	wTimeout time.Duration
}

func (c *connection) reset() {
	c.id = ""
	c.conn = nil
	c.rTimeout = 0
	c.wTimeout = 0
}

func (c *connection) read() ([]byte, error) {
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

func (c *connection) write(data []byte) error {
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *connection) writeText(text string) error {
	return c.conn.WriteMessage(websocket.TextMessage, utils.StringToBytes(text))
}

func (c *connection) close() error {
	return c.conn.Close()
}

func (c *connection) writeAndClose(text string) error {
	errW := c.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseInternalServerErr, text))
	errC := c.conn.Close()
	return errors.Join(errW, errC)
}

func (c *connection) setReadTimeout(timeout time.Duration) error {
	return c.conn.SetReadDeadline(time.Now().Add(timeout))
}

func (c *connection) setWriteTimeout(timeout time.Duration) error {
	return c.conn.SetWriteDeadline(time.Now().Add(timeout))
}

func (c *connection) getId() string {
	return c.id
}

func (c *connection) getRemote() string {
	return c.conn.RemoteAddr().String()
}
