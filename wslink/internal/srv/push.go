package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
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
	userConns, err := s.sessBiz.GetSessionByUidDevice(ctx, uid, device)
	if err != nil {
		return xerror.Wrapf(err, "sess failed to get user conns").WithCtx(ctx)
	}

	if len(userConns) == 0 {
		// not online
		return nil
	}

	// TODO 判断哪些连接是在本机，哪些连接不在本机
	for _, uConn := range userConns {
		uConn.Send(ctx, data)
	}

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
