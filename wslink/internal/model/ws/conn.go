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

// websocket连接
type connection struct {
	id   string
	conn *websocket.Conn

	rTimeout time.Duration
	wTimeout time.Duration
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

func (c *connection) reset() {
	c.id = ""
	c.conn = nil
	c.rTimeout = 0
	c.wTimeout = 0
}

func (c *connection) read() ([]byte, error) {
	c.setReadTimeout(c.rTimeout)
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
		return nil, ErrConnClosed
	}

	// binary data
	return data, err
}

func (c *connection) write(data []byte) error {
	c.setWriteTimeout(c.wTimeout)
	return c.conn.WriteMessage(websocket.BinaryMessage, data)
}

func (c *connection) writeText(text string) error {
	c.setWriteTimeout(c.wTimeout)
	return c.conn.WriteMessage(websocket.TextMessage, utils.StringToBytes(text))
}

func (c *connection) close() error {
	return c.conn.Close()
}

func (c *connection) graceClose() error {
	err := c.conn.WriteControl(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, "CLOSED"),
		time.Time{})
	err2 := c.conn.Close()
	return errors.Join(err, err2)
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
