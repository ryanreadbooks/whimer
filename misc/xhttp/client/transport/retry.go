package transport

import (
	"net/http"
	"time"
)

// RetryOptions 重试配置
type RetryOptions struct {
	// 最大重试次数
	MaxAttempts int
	// 初始退避时间
	InitialBackoff time.Duration
	// 最大退避时间
	MaxBackoff time.Duration
	// 退避倍数
	Multiplier float64
	// 判断是否需要重试的函数
	ShouldRetry func(resp *http.Response, err error) bool
}

// DefaultRetryOptions 默认重试配置
func DefaultRetryOptions() RetryOptions {
	return RetryOptions{
		MaxAttempts:    3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		Multiplier:     2,
		ShouldRetry:    DefaultShouldRetry,
	}
}

// DefaultShouldRetry 默认重试判断逻辑
func DefaultShouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	// 5xx 服务端错误重试
	if resp != nil && resp.StatusCode >= 500 {
		return true
	}
	return false
}

// Retry 创建重试 transport
func Retry(opts RetryOptions, next http.RoundTripper) http.RoundTripper {
	if opts.MaxAttempts <= 0 {
		opts.MaxAttempts = 3
	}
	if opts.InitialBackoff <= 0 {
		opts.InitialBackoff = 100 * time.Millisecond
	}
	if opts.MaxBackoff <= 0 {
		opts.MaxBackoff = 5 * time.Second
	}
	if opts.Multiplier <= 0 {
		opts.Multiplier = 2
	}
	if opts.ShouldRetry == nil {
		opts.ShouldRetry = DefaultShouldRetry
	}

	return Transporter(func(req *http.Request) (*http.Response, error) {
		var resp *http.Response
		var err error
		backoff := opts.InitialBackoff

		for attempt := 1; attempt <= opts.MaxAttempts; attempt++ {
			resp, err = next.RoundTrip(req)

			// 成功或不需要重试
			if !opts.ShouldRetry(resp, err) {
				return resp, err
			}

			// 最后一次尝试，直接返回
			if attempt == opts.MaxAttempts {
				return resp, err
			}

			// 等待退避时间
			select {
			case <-req.Context().Done():
				if err == nil {
					err = req.Context().Err()
				}
				return resp, err
			case <-time.After(backoff):
			}

			// 计算下次退避时间
			backoff = time.Duration(float64(backoff) * opts.Multiplier)
			if backoff > opts.MaxBackoff {
				backoff = opts.MaxBackoff
			}
		}

		return resp, err
	})
}

// RetryWithContext 带 context 感知的重试
func RetryWithContext(opts RetryOptions, next http.RoundTripper) http.RoundTripper {
	return Retry(opts, next)
}
