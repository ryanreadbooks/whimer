package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dep"
	pushv1 "github.com/ryanreadbooks/whimer/wslink/api/push/v1"
)

// TODO optimize 先hold 再批量通知
func (s *P2PChatSrv) notifyReceiver(ctx context.Context, receiver int64) {
	// 下发通知
	_, err := dep.Notifier().Broadcast(ctx, &pushv1.BroadcastRequest{
		Targets: []int64{receiver},
		Data:    []byte("MSGER"), // TODO 精细化通知下发, 区分p2p和system
	})
	if err != nil {
		xlog.Msgf("p2p chat failed to notify user %d", receiver).Err(err).Errorx(ctx)
	}
}
