package pushcenter

import (
	"context"
	"encoding/json"
)

// websocket 推送 cmd 定义
type cmd string

const (
	cmdWhisperMsgNotify cmd = "whisper_notify"
	cmdSysMsgNotify     cmd = "sys_notify"
)

// websocket 推送 action 定义
type action string

const (
	actionPullWhisper action = "pull_whisper"
	actionPullUnreads action = "pull_unreads"
)

type cmdAction struct {
	Cmd     cmd      `json:"cmd"`
	Actions []action `json:"actions"`
}

func newCmdAction(c cmd, a action, actions ...action) cmdAction {
	return cmdAction{
		Cmd:     c,
		Actions: append([]action{a}, actions...),
	}
}

func (c cmdAction) bytes() []byte {
	b, _ := json.Marshal(c)
	return b
}

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
	data := newCmdAction(cmdSysMsgNotify, actionPullUnreads).bytes()
	return pusher.Broadcast(ctx, recvUids, data)
}

func NotifyWhisperMsg(ctx context.Context, recvUid int64) error {
	return BatchNotifyWhisperMsg(ctx, []int64{recvUid})
}

func BatchNotifyWhisperMsg(ctx context.Context, recvUids []int64) error {
	if len(recvUids) == 0 {
		return nil
	}
	data := newCmdAction(cmdWhisperMsgNotify, actionPullWhisper).bytes()
	return pusher.Broadcast(ctx, recvUids, data)
}
