package pushcenter

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dep"
	pushv1 "github.com/ryanreadbooks/whimer/wslink/api/push/v1"
)

type WsPusher struct{}

func NewWsPusher() *WsPusher {
	return &WsPusher{}
}

func (p *WsPusher) Broadcast(ctx context.Context, targets []int64, data []byte) error {
	_, err := dep.WebsocketPusher().Broadcast(ctx, &pushv1.BroadcastRequest{
		Targets: targets,
		Data:    data,
	})
	return err
}
