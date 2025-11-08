package pushcenter

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/pushcmd"
	pushv1 "github.com/ryanreadbooks/whimer/wslink/api/push/v1"
)

func NotifyWhisperMsg(ctx context.Context, recvUid int64) error {
	data := pushcmd.NewCmdAction(pushcmd.CmdWhisperMsgNotify, pushcmd.ActionPullWhisper).Bytes()
	_, err := dep.WebsocketPusher().Broadcast(ctx, &pushv1.BroadcastRequest{
		Targets: []int64{recvUid},
		Data:    data,
	})

	return err
}
