package srv

import (
	"context"
	"sync"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/wslink/internal/biz"
	"github.com/ryanreadbooks/whimer/wslink/internal/model"
)

type PushService struct {
	sessBiz biz.SessionBiz

	frMu       sync.RWMutex
	forwarders map[string]*Forwarder
}

func NewPushService(b biz.Biz) *PushService {
	return &PushService{
		sessBiz:    b.SessionBiz,
		forwarders: make(map[string]*Forwarder),
	}
}

// 给特定uid的特定device设备发送data
func (s *PushService) Push(ctx context.Context, uid int64, device model.Device, data []byte) error {
	locals, nonLocals, err := s.sessBiz.RespectivelyGetSessionByUids(ctx, []int64{uid})
	if err != nil {
		return xerror.Wrapf(err, "failed to get session by uid").WithCtx(ctx)
	}

	concurrent.SafeGo(func() {
		err := s.PushLocalConns(ctx, FormatPushLocalConnReq(locals, device, data))
		if err != nil {
			xlog.Msgf("push local conns err").Err(err).Errorx(ctx)
		}
	})
	concurrent.SafeGo(func() {
		err := s.PushNonLocalConns(ctx, FormatPushNonLocalConnReq(nonLocals, device, data))
		if err != nil {
			xlog.Msgf("push non local conns err").Err(err).Errorx(ctx)
		}
	})

	return nil
}

func (s *PushService) PushLocalConns(ctx context.Context, datas []*PushLocalConnReq) error {
	if len(datas) == 0 {
		return nil
	}

	ctx = context.WithoutCancel(ctx)
	// 批量推送conns
	var wg sync.WaitGroup
	err := xslice.BatchAsyncExec(&wg, datas, 100, func(start, end int) error {
		cs := datas[start:end]
		for _, c := range cs {
			if err := c.Conn.Send(ctx, c.Data); err != nil {
				xlog.Msgf("push local conn %s err", c.Conn.GetId()).Err(err).Errorx(ctx)
			}
		}

		return nil
	})
	if err != nil {
		return xerror.Wrapf(err, "push local conns err").WithCtx(ctx)
	}

	return nil
}

func (s *PushService) PushNonLocalConns(ctx context.Context, conns []*PushNonLocalConnReq) error {
	if len(conns) == 0 {
		return nil
	}

	ctx = context.WithoutCancel(ctx)

	var totalCnt = len(conns)
	var successCnt = 0

	defer func() {
		xlog.Msgf("push non local sucess/total(%d/%d)", successCnt, totalCnt).Debugx(ctx)
	}()

	// addr -> []Conns
	m := make(map[string][]*PushNonLocalConnReq)
	for _, c := range conns {
		m[c.Conn.GetLocalIp()] = append(m[c.Conn.GetLocalIp()], c)
	}

	// 转发
	failure := make(map[string][]*PushNonLocalConnReq)
	success := make(map[string][]*PushNonLocalConnReq)
	for target, conns := range m {
		forward, err := s.GetForwarder(ctx, target)
		if err != nil {
			// 拿不到forwarder, 放入失败，给成功的重试
			xlog.Msgf("push failed to get forward at %s", target).Err(err).Errorx(ctx)
			failure[target] = append(failure[target], conns...)
			continue
		}

		err = forward.Forward(ctx, conns)
		if err != nil {
			xlog.Msgf("push failed to forward to %s", target).Err(err).Errorx(ctx)
			failure[target] = append(failure[target], conns...)
		} else {
			success[target] = append(success[target], conns...)
			successCnt += len(conns)
		}
	}

	// 将失败转发给成功的forward 尽量重试
	if len(success) == 0 {
		return xerror.ErrInternal.Msg("all non local forward failed")
	}

	for successTarget := range success {
		forwarder, err := s.GetForwarder(ctx, successTarget)
		if err != nil {
			xlog.Msgf("retry push failed to get forward at %s", successTarget).Err(err).Errorx(ctx)
			continue
		}

		for failTarget, conns := range failure {
			err = forwarder.Forward(ctx, conns)
			if err != nil {
				xlog.Msgf("retry push failed to forward to %s by %s", failTarget, successTarget).
					Err(err).Errorx(ctx)
			}
			successCnt += len(conns)
		}
		break
	}

	return nil
}

type BatchPushReq struct {
	Uid    int64
	Device model.Device
	Data   []byte
}

// 广播给uids下所有设备相同的data
func (s *PushService) Broadcast(ctx context.Context, uids []int64, data []byte) error {
	if len(uids) == 0 {
		return nil
	}
	locals, nonLocals, err := s.sessBiz.RespectivelyGetSessionByUids(ctx, uids)
	if err != nil {
		return xerror.Wrapf(err, "failed to get session by uid").WithCtx(ctx)
	}

	concurrent.SafeGo(func() {
		err := s.PushLocalConns(ctx, FormatPushLocalConnReq(locals, "", data))
		if err != nil {
			xlog.Msgf("broadcast local conns err").Err(err).Errorx(ctx)
		}
	})
	concurrent.SafeGo(func() {
		err := s.PushNonLocalConns(ctx, FormatPushNonLocalConnReq(nonLocals, "", data))
		if err != nil {
			xlog.Msgf("broadcast non local conns err").Err(err).Errorx(ctx)
		}
	})

	return nil
}

func (s *PushService) BatchPush(ctx context.Context, reqs []BatchPushReq) error {
	// 直接分批调用Push
	concurrent.SafeGo(func() {
		var wg sync.WaitGroup
		ctx = context.WithoutCancel(ctx)
		xslice.BatchAsyncExec(&wg, reqs, 100, func(start, end int) error {
			for _, req := range reqs[start:end] {
				s.Push(ctx, req.Uid, req.Device, req.Data)
			}
			return nil
		})
	})

	return nil
}

func (s *PushService) GetForwarder(ctx context.Context, target string) (*Forwarder, error) {
	s.frMu.RLock()

	var forwarder *Forwarder
	var ok bool

	forwarder, ok = s.forwarders[target]
	if !ok {
		s.frMu.RUnlock()
		// 不存在需要创建
		var newForwarder *Forwarder
		for idx := range 3 {
			fdr, err := NewForwarder(target)
			if err != nil {
				xlog.Msgf("push failed to new forwarder at %s (retry = %d)", target, idx).Err(err).Errorx(ctx)
				time.Sleep(time.Second)
				continue
			}
			newForwarder = fdr
			break
		}

		s.frMu.Lock()
		// check again
		if curForwarder, ok := s.forwarders[target]; ok {
			forwarder = curForwarder // we already have this
			newForwarder.Close()     // close this newly created forwarder
		} else {
			s.forwarders[target] = newForwarder
			forwarder = newForwarder
			xlog.Msgf("push created new forwarder at %s", target).Infox(ctx)
		}
		s.frMu.Unlock()
	} else {
		s.frMu.RUnlock()
	}

	return forwarder, nil
}
