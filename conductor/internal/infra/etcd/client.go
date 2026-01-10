package etcd

import (
	"fmt"

	"github.com/ryanreadbooks/whimer/conductor/internal/config"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Client struct {
	client *clientv3.Client
}

func New(c *config.Config) (*Client, error) {
	host := c.Etcd.Hosts

	client, err := clientv3.New(clientv3.Config{Endpoints: host})
	if err != nil {
		return nil, err
	}

	return &Client{
		client: client,
	}, nil
}

func MustNew(c *config.Config) *Client {
	client, err := New(c)
	if err != nil {
		panic(fmt.Errorf("etcd client init failed: %w", err))
	}

	return client
}

func (c *Client) GetClient() *clientv3.Client {
	return c.client
}
