package push

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/pushcmd"
	pushv1 "github.com/ryanreadbooks/whimer/wslink/api/push/v1"
)

// 推送相关功能封装

func PushSysCmdPullUnreadAction(ctx context.Context, recvUid int64) error {
	return BatchPushSysCmdPullUnreadAction(ctx, []int64{recvUid})
}

func BatchPushSysCmdPullUnreadAction(ctx context.Context, recvUids []int64) error {
	data := pushcmd.NewCmdAction(pushcmd.CmdSysMsgNotify, pushcmd.ActionPullUnreads).Bytes()
	_, err := dep.WebsocketPusher().Broadcast(ctx, &pushv1.BroadcastRequest{
		Targets: recvUids,
		Data:    data,
	})

	return err
}

func PushP2PCmdPullP2PAction(ctx context.Context, recvUid int64) error {
	data := pushcmd.NewCmdAction(pushcmd.CmdP2PMsgNotify, pushcmd.ActionPullP2P).Bytes()
	_, err := dep.WebsocketPusher().Broadcast(ctx, &pushv1.BroadcastRequest{
		Targets: []int64{recvUid},
		Data:    data,
	})

	return err
}
