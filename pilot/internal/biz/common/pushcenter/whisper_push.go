package pushcenter

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dep"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/pushcmd"
	pushv1 "github.com/ryanreadbooks/whimer/wslink/api/push/v1"
)

func NotifyWhisperMsg(ctx context.Context, recvUid int64) error {
	return BatchNotifyWhisperMsg(ctx, []int64{recvUid})
}

func BatchNotifyWhisperMsg(ctx context.Context, recvUids []int64) error {
	if len(recvUids) == 0 {
		return nil
	}
	
	data := pushcmd.NewCmdAction(pushcmd.CmdWhisperMsgNotify, pushcmd.ActionPullWhisper).Bytes()
	_, err := dep.WebsocketPusher().Broadcast(ctx, &pushv1.BroadcastRequest{
		Targets: recvUids,
		Data:    data,
	})

	return err
}
