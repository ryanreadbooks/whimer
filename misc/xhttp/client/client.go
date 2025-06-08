package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xhttp/client/transport"
)

// 内置一些通用的RoundTripper的http client
// 比如超时、重试、熔断等
type Client struct {
	impl *http.Client
}

func New(schema, host string) *Client {
	var tp = http.DefaultTransport
	tp = transport.AttachHostPrefix(schema, host, tp)

	return &Client{
		impl: &http.Client{
			Transport: tp,
		},
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.impl.Do(req)
}

func (c *Client) Fetch(req *http.Request, output any) ([]byte, error) {
	res, err := c.impl.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request failed with status: %s: %s", res.Status, body)
	}

	// try unmarshal the body into the output
	err = json.Unmarshal(body, output)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	return body, nil
}
