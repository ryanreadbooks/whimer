package queue

import (
	"context"

	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/comment/internal/repo/comm"
)

type Bus struct {
	topic string
}

func New(c *config.Config) *Bus {
	return &Bus{}
}

const (
	ActAddReply = 1 + iota
	ActDelReply
)

func (b *Bus) pushReplyAct(ctx context.Context) error {
	return nil
}

func (b *Bus) AddReply(ctx context.Context, data *comm.Model) error {

	return nil
}

func (b *Bus) DelReply(ctx context.Context, data *comm.Model) error {

	return nil
}
