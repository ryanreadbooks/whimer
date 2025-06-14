package grpc

import (
	"context"

	pushv1 "github.com/ryanreadbooks/whimer/wslink/api/push/v1"
	"github.com/ryanreadbooks/whimer/wslink/internal/global"
	"github.com/ryanreadbooks/whimer/wslink/internal/model"
	"github.com/ryanreadbooks/whimer/wslink/internal/srv"
)

type PushServiceServer struct {
	pushv1.UnimplementedPushServiceServer

	Svc *srv.Service
}

func NewPushServiceServer(ctx *srv.Service) *PushServiceServer {
	return &PushServiceServer{
		Svc: ctx,
	}
}

// 给uid推送数据
func (s *PushServiceServer) Push(ctx context.Context, in *pushv1.PushRequest) (*pushv1.PushResponse, error) {
	if in.Uid == 0 {
		return nil, global.ErrUserEmpty
	}
	device := model.GetDeviceFromPb(in.Device)
	if device.Empty() || device.Unspec() {
		return nil, global.ErrUnsupportedDevice
	}
	if len(in.Data) == 0 {
		return nil, global.ErrDataEmpty
	}

	err := s.Svc.PushService.Push(ctx, in.Uid, device, in.Data)
	if err != nil {
		return nil, err
	}

	return &pushv1.PushResponse{}, nil
}

// 消息广播
func (s *PushServiceServer) Broadcast(ctx context.Context, in *pushv1.BroadcastRequest) (
	*pushv1.BroadcastResponse, error) {
	if len(in.Targets) == 0 {
		return nil, global.ErrUserEmpty
	}
	if len(in.GetData()) == 0 {
		return nil, global.ErrDataEmpty
	}

	err := s.Svc.PushService.Broadcast(ctx, in.Targets, in.Data)
	if err != nil {
		return nil, err
	}

	return &pushv1.BroadcastResponse{}, nil
}

// 批量推送 每个用户推送的数据不一样
func (s *PushServiceServer) BatchPush(ctx context.Context, in *pushv1.BatchPushRequest) (
	*pushv1.BatchPushResponse, error) {

	if len(in.Targets) == 0 {
		return nil, global.ErrUserEmpty
	}

	reqs := make([]srv.BatchPushReq, 0, len(in.Targets))
	for _, r := range in.Targets {
		device := model.GetDeviceFromPb(r.Device)
		if r.Uid != 0 && len(r.Data) > 0 && !device.Empty() && !device.Unspec() {
			reqs = append(reqs, srv.BatchPushReq{
				Uid:    r.Uid,
				Data:   r.Data,
				Device: device,
			})
		}
	}

	if len(reqs) == 0 {
		return nil, global.ErrUserEmpty
	}

	err := s.Svc.PushService.BatchPush(ctx, reqs)
	if err != nil {
		return nil, err
	}

	return &pushv1.BatchPushResponse{}, nil
}
