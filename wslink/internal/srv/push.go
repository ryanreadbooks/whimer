package srv

import (
	"context"
	"sync"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/wslink/internal/biz"
	"github.com/ryanreadbooks/whimer/wslink/internal/model"
)

type PushService struct {
	sessBiz biz.SessionBiz
}

func NewPushService(b biz.Biz) *PushService {
	return &PushService{
		sessBiz: b.SessionBiz,
	}
}

func (s *PushService) Push(ctx context.Context, uid int64, device model.Device, data []byte) error {
	locals, nonLocals, err := s.sessBiz.RespectivelyGetSessionByUid(ctx, uid)
	if err != nil {
		return xerror.Wrapf(err, "failed to get session by uid").WithCtx(ctx)
	}

	// filter devices
	localTarges := make([]*pushConnsReq, 0, len(locals))
	for _, l := range locals {
		if l.GetDevice() == device {
			localTarges = append(localTarges, &pushConnsReq{
				conn: l,
				data: data,
			})
		}
	}
	nonLocalTargets := make([]*pushUnsendConnReq, 0, len(nonLocals))
	for _, nl := range nonLocals {
		if nl.GetDevice() == device {
			nonLocalTargets = append(nonLocalTargets, &pushUnsendConnReq{
				conn: nl,
				data: data,
			})
		}
	}

	concurrent.SafeGo(func() {
		err := s.pushLocalConns(ctx, localTarges)
		if err != nil {
			xlog.Msgf("push local conns err").Err(err).Errorx(ctx)
		}
	})
	concurrent.SafeGo(func() {
		err := s.pushNonLocalConns(ctx, nonLocalTargets)
		if err != nil {
			xlog.Msgf("push non local conns err").Err(err).Errorx(ctx)
		}
	})

	return nil
}

type pushConnsReq struct {
	conn biz.Session
	data []byte
}

type pushUnsendConnReq struct {
	conn biz.UnSendableSession
	data []byte
}

func (s *PushService) pushLocalConns(ctx context.Context, datas []*pushConnsReq) error {
	if len(datas) == 0 {
		return nil
	}

	ctx = context.WithoutCancel(ctx)
	// 批量推送conns
	var wg sync.WaitGroup
	err := xslice.BatchAsyncExec(&wg, datas, 100, func(start, end int) error {
		cs := datas[start:end]
		for _, c := range cs {
			if err := c.conn.Send(ctx, c.data); err != nil {
				xlog.Msgf("push local conn %s err", c.conn.GetId()).Err(err).Errorx(ctx)
			}
		}

		return nil
	})
	if err != nil {
		return xerror.Wrapf(err, "push local conns err").WithCtx(ctx)
	}

	return nil
}

func (s *PushService) pushNonLocalConns(ctx context.Context, conns []*pushUnsendConnReq) error {

	return nil
}

type BatchPushReq struct {
	Uid    int64
	Device model.Device
	Data   []byte
}

func (s *PushService) Broadcast(ctx context.Context, uids []int64, data []byte) error {
	return nil
}

func (s *PushService) BatchPush(ctx context.Context, reqs []BatchPushReq) error {

	return nil
}
