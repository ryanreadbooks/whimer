package pushcenter

import (
	"context"

	pushv1 "github.com/ryanreadbooks/whimer/idl/gen/go/wslink/api/push/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dep"
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
