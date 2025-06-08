package srv

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	forwardv1 "github.com/ryanreadbooks/whimer/wslink/api/forward/v1"
	"github.com/ryanreadbooks/whimer/wslink/internal/biz"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
)

const (
	maxForwardCntAllowed = 5
)

// 同服务不同实例之间转发(grpc)
type Forwarder struct {
	cc   *grpc.ClientConn
	impl forwardv1.ForwardServiceClient
}

func NewForwarder(addr string) (*Forwarder, error) {
	fr := &Forwarder{}
	c, err := zrpc.NewClientWithTarget(addr,
		zrpc.WithUnaryClientInterceptor(interceptor.UnaryClientErrorHandler),
		zrpc.WithUnaryClientInterceptor(interceptor.UnaryClientMetadataInject),
		zrpc.WithDialOption(grpc.WithConnectParams(
			grpc.ConnectParams{
				Backoff:           backoff.DefaultConfig,
				MinConnectTimeout: 8 * time.Second,
			},
		)),
	)
	if err != nil {
		return nil, err
	}

	fr.cc = c.Conn()
	fr.impl = forwardv1.NewForwardServiceClient(c.Conn())

	return fr, nil
}

// 批量转发
func (f *Forwarder) Forward(ctx context.Context, conns []*PushNonLocalConnReq) error {
	targets := make([]*forwardv1.ForwardTarget, 0, len(conns))
	for _, c := range conns {
		if c != nil {
			targets = append(targets, &forwardv1.ForwardTarget{
				Id:         c.Conn.GetId(),
				Data:       c.Data,
				ForwardCnt: int32(c.ForwardCnt),
			})
		}
	}
	in := forwardv1.PushForwardRequest{Targets: targets}
	_, err := f.impl.PushForward(ctx, &in)
	return err
}

func (f *Forwarder) Close() {
	f.cc.Close()
}

type ForwardService struct {
	sessBiz     biz.SessionBiz
	pushService *PushService
}

func NewForwardService(b biz.Biz, pushServ *PushService) *ForwardService {
	f := ForwardService{
		sessBiz:     b.SessionBiz,
		pushService: pushServ,
	}

	return &f
}

type ForwardReq struct {
	SessId     string
	Data       []byte
	ForwardCnt int32
}

func (s *ForwardService) Forward(ctx context.Context, reqs []*ForwardReq) error {
	// 再次检查sessId是否为本机的连接
	reqMap := make(map[string]*ForwardReq, len(reqs))
	sessIds := make([]string, 0, len(reqs))
	for _, req := range reqs {
		sessIds = append(sessIds, req.SessId)
		reqMap[req.SessId] = req
	}

	locals, nonLocals, err := s.sessBiz.RespectivelyGetSessionByIds(ctx, sessIds)
	if err != nil {
		return xerror.Wrapf(err, "forward failed to get session by uid").WithCtx(ctx)
	}

	// locals可以直接发送
	localTargets := make([]*PushLocalConnReq, 0, len(locals))
	for _, l := range locals {
		if l != nil {
			localTargets = append(localTargets, &PushLocalConnReq{
				Conn: l,
				Data: reqMap[l.GetId()].Data,
			})
		}
	}

	s.pushService.PushLocalConns(ctx, localTargets)

	// 这部分需要再次转发
	nonLocalTargets := make([]*PushNonLocalConnReq, 0, len(nonLocals))
	for _, nl := range nonLocals {
		if nl != nil {
			curForwardCnt := reqMap[nl.GetId()].ForwardCnt
			if curForwardCnt > maxForwardCntAllowed {
				xlog.Msgf("session %s reached max forward cnt, drop it", nl.GetId()).Errorx(ctx)
				continue
			}

			nonLocalTargets = append(nonLocalTargets, &PushNonLocalConnReq{
				Conn:       nl,
				Data:       reqMap[nl.GetId()].Data,
				ForwardCnt: int(curForwardCnt) + 1, // 增加一次转发
			})
		}
	}
	s.pushService.PushNonLocalConns(ctx, nonLocalTargets)

	return nil
}
