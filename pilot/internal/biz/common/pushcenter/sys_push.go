package pushcenter

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/pushcmd"
	pushv1 "github.com/ryanreadbooks/whimer/wslink/api/push/v1"
)

// 推送相关功能封装

func batchPushSysMsg(ctx context.Context, recvUids []int64) error {
	if len(recvUids) == 0 {
		return nil
	}

	data := pushcmd.NewCmdAction(pushcmd.CmdSysMsgNotify, pushcmd.ActionPullUnreads).Bytes()
	_, err := dep.WebsocketPusher().Broadcast(ctx, &pushv1.BroadcastRequest{
		Targets: recvUids,
		Data:    data,
	})

	return err
}

func NotifySystemMsg(ctx context.Context, recvUid int64) error {
	return BatchNotifySystemMsg(ctx, []int64{recvUid})
}

func BatchNotifySystemMsg(ctx context.Context, recvUids []int64) error {
	return batchPushSysMsg(ctx, recvUids)
}
