package client

// 内置一些通用的RoundTripper的http client
// 比如超时、重试、熔断等
type Client struct{}

func New() *Client {
	return &Client{}
}
