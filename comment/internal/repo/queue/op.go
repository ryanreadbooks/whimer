package queue

import (
	"context"

	"github.com/ryanreadbooks/whimer/comment/internal/repo/comm"
)

type Bus struct {

}

func (b *Bus) pushReplyAct(ctx context.Context) error {
	return nil
}

func (b *Bus) AddReply(ctx context.Context, data *comm.Model) error {

	return nil
}
