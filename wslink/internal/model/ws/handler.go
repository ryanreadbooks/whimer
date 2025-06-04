package ws

import "context"

// 数据上行处理
type SessionOnDataHandler interface {
	OnData(ctx context.Context, s *Session, data []byte) error
}

// session关闭处理
type SessionOnClosedHandler interface {
	OnClosed(ctx context.Context, s *Session) error
}

// session创建时处理
type SessionOnCreateHandler interface {
	// 确认是否可以建立连接，返回非空error表明不能建立连接
	OnCreate(ctx context.Context, s *Session) error
}

// session的处理接口
type SessionHandler interface {
	SessionOnCreateHandler
	SessionOnDataHandler
	SessionOnClosedHandler
}
