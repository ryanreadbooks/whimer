package client

import "net/http"

// 内置一些通用的RoundTripper的http client
// 比如超时、重试、熔断等
type Client struct {
	impl *http.Client
}

func New() *Client {
	cli := &http.Client{}

	return &Client{
		impl: cli,
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.impl.Do(req)
}
