package ws

import "context"

// 数据上行处理
type ConnectionOnDataHandler interface {
	OnData(ctx context.Context, c *Connection, data []byte) error
}

// session关闭处理
type ConnectionAfterClosedHandler interface {
	AfterClosed(ctx context.Context, id string) error
}

// session创建时处理
type ConnectionOnCreateHandler interface {
	// 确认是否可以建立连接，返回非空error表明不能建立连接
	OnCreate(ctx context.Context, c *Connection) error
}

// session的处理接口
type ConnectionHandler interface {
	ConnectionOnCreateHandler
	ConnectionOnDataHandler
	ConnectionAfterClosedHandler
}
