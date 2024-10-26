package global

// 计数服务的业务码
const (
	CommentLikeBizcode    int32 = 40001 + iota // 点赞
	CommentDislikeBizcode                      // 点踩
)

// 代理模式
type ProxyMode string

const (
	ProxyModeBus    ProxyMode = "bus"
	ProxyModeDirect ProxyMode = "direct"
)
