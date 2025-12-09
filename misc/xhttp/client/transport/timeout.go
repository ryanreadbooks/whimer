package transport

import (
	"context"
	"net/http"
	"time"
)

// Timeout 创建超时控制 transport
func Timeout(timeout time.Duration, next http.RoundTripper) http.RoundTripper {
	return Transporter(func(req *http.Request) (*http.Response, error) {
		if timeout <= 0 {
			return next.RoundTrip(req)
		}

		ctx, cancel := context.WithTimeout(req.Context(), timeout)
		defer cancel()

		req = req.WithContext(ctx)
		return next.RoundTrip(req)
	})
}
