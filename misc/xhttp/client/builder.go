package client

import (
	"net/http"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xhttp/client/transport"
)

// Builder HTTP 客户端构建器
type Builder struct {
	baseTransport http.RoundTripper
	transports    []func(http.RoundTripper) http.RoundTripper
}

// NewBuilder 创建客户端构建器
func NewBuilder() *Builder {
	return &Builder{
		baseTransport: http.DefaultTransport,
		transports:    make([]func(http.RoundTripper) http.RoundTripper, 0),
	}
}

// WithBaseTransport 设置基础 transport
func (b *Builder) WithBaseTransport(t http.RoundTripper) *Builder {
	b.baseTransport = t
	return b
}

// WithTransport 添加自定义 transport（按添加顺序包装）
func (b *Builder) WithTransport(wrapper func(http.RoundTripper) http.RoundTripper) *Builder {
	b.transports = append(b.transports, wrapper)
	return b
}

// WithTimeout 添加超时控制
func (b *Builder) WithTimeout(timeout time.Duration) *Builder {
	b.transports = append(b.transports, func(next http.RoundTripper) http.RoundTripper {
		return transport.Timeout(timeout, next)
	})
	return b
}

// WithRetry 添加重试机制
func (b *Builder) WithRetry(opts transport.RetryOptions) *Builder {
	b.transports = append(b.transports, func(next http.RoundTripper) http.RoundTripper {
		return transport.Retry(opts, next)
	})
	return b
}

// WithDefaultRetry 添加默认重试机制
func (b *Builder) WithDefaultRetry() *Builder {
	return b.WithRetry(transport.DefaultRetryOptions())
}

// WithTracing 添加链路追踪
func (b *Builder) WithTracing() *Builder {
	b.transports = append(b.transports, func(next http.RoundTripper) http.RoundTripper {
		return transport.SpanTracing(next)
	})
	return b
}

// WithHostPrefix 添加 host 前缀
func (b *Builder) WithHostPrefix(schema, host string) *Builder {
	b.transports = append(b.transports, func(next http.RoundTripper) http.RoundTripper {
		return transport.AttachHostPrefix(schema, host, next)
	})
	return b
}

// Build 构建 HTTP 客户端
// transport 包装顺序：最后添加的在最外层
// 例如：WithTimeout -> WithRetry -> WithTracing
// 执行顺序：Tracing -> Retry -> Timeout -> BaseTransport
func (b *Builder) Build() *http.Client {
	rt := b.baseTransport

	// 逆序包装，最后添加的在最外层
	for i := len(b.transports) - 1; i >= 0; i-- {
		rt = b.transports[i](rt)
	}

	return &http.Client{
		Transport: rt,
	}
}

// BuildTransport 只构建 transport，不创建 client
func (b *Builder) BuildTransport() http.RoundTripper {
	rt := b.baseTransport

	for i := len(b.transports) - 1; i >= 0; i-- {
		rt = b.transports[i](rt)
	}

	return rt
}
