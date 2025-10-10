package websocket

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dep"
	wspushv1 "github.com/ryanreadbooks/whimer/wslink/api/push/v1"
)

type Biz struct{}

func NewWebsocketBiz() Biz {
	return Biz{}
}

// 通知p2p消息
func (b *Biz) NotifyUserMsg(ctx context.Context, receiver int64) error {
	_, err := dep.WsLinker().Broadcast(ctx, &wspushv1.BroadcastRequest{
		Targets: []int64{receiver},
		Data:    []byte(NotifyUserMsg),
	})
	if err != nil {
		return xerror.Wrapf(err, "websocket failed to notify user %d", receiver).WithCtx(ctx)
	}

	return nil
}

func (b *Biz) NotifySysNotice(ctx context.Context, receiver int64) error {

	return nil
}

func (b *Biz) NotifySysReply(ctx context.Context, receiver int64) error {

	return nil
}

// 通知被@的人
func (b *Biz) NotifySysMention(ctx context.Context, receiver int64, targets []*NotifySysContent) error {
	var errs []error
	for _, target := range targets {
		content := &notifySysContentInner{
			Type:             NotifySysMention,
			NotifySysContent: target,
		}
		data, err := json.Marshal(content)
		if err == nil {
			if _, err := dep.WsLinker().Broadcast(ctx, &wspushv1.BroadcastRequest{
				Targets: []int64{receiver},
				Data:    data,
			}); err != nil {
				errs = append(errs, xerror.Wrapf(err, "websocket failed to notify user %d", receiver).WithCtx(ctx))
			}
		}
	}

	return errors.Join(errs...)
}

func (b *Biz) NotifySysLikes(ctx context.Context, receiver int64) error {

	return nil
}
