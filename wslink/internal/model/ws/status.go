package ws

type SessionStatus string

const (
	StatusNoActive SessionStatus = "noactive" // 关闭
	StatusActive   SessionStatus = "active"   // 获取中
	StatusPending  SessionStatus = "pending"  // wslink服务重启过程
)

// implements encoding.BinaryMarshaler
func (d SessionStatus) MarshalBinary() ([]byte, error) {
	return []byte(d), nil
}
