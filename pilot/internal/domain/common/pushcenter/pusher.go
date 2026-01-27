package pushcenter

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/model/pushcmd"
)

type Pusher interface {
	Broadcast(ctx context.Context, targets []int64, data []byte) error
}

var pusher Pusher

func SetPusher(p Pusher) {
	pusher = p
}

func NotifySystemMsg(ctx context.Context, recvUid int64) error {
	return BatchNotifySystemMsg(ctx, []int64{recvUid})
}

func BatchNotifySystemMsg(ctx context.Context, recvUids []int64) error {
	if len(recvUids) == 0 {
		return nil
	}
	data := pushcmd.NewCmdAction(pushcmd.CmdSysMsgNotify, pushcmd.ActionPullUnreads).Bytes()
	return pusher.Broadcast(ctx, recvUids, data)
}

func NotifyWhisperMsg(ctx context.Context, recvUid int64) error {
	return BatchNotifyWhisperMsg(ctx, []int64{recvUid})
}

func BatchNotifyWhisperMsg(ctx context.Context, recvUids []int64) error {
	if len(recvUids) == 0 {
		return nil
	}
	data := pushcmd.NewCmdAction(pushcmd.CmdWhisperMsgNotify, pushcmd.ActionPullWhisper).Bytes()
	return pusher.Broadcast(ctx, recvUids, data)
}
