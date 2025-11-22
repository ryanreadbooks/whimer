package ws

type SessionStatus string

const (
	StatusNoActive SessionStatus = "noactive" // 关闭
	StatusActive   SessionStatus = "active"   // 活跃中
	StatusPending  SessionStatus = "pending"  // wslink服务重启过程
)

func (s SessionStatus) NoActive() bool {
	return s == StatusNoActive
}

func (s SessionStatus) Active() bool {
	return s == StatusActive
}

func (s SessionStatus) Pending() bool {
	return s == StatusPending
}

// implements encoding.BinaryMarshaler
func (d SessionStatus) MarshalBinary() ([]byte, error) {
	return []byte(d), nil
}
