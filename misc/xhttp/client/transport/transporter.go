package transport

import "net/http"

type Transporter func(req *http.Request) (*http.Response, error)

// 实现http.RroundTripper接口
func (t Transporter) RoundTrip(req *http.Request) (*http.Response, error) {
	return t(req)
}
