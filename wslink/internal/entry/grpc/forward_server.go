package grpc

import (
	"context"

	forwardv1 "github.com/ryanreadbooks/whimer/wslink/api/forward/v1"
	"github.com/ryanreadbooks/whimer/wslink/internal/srv"
)

type ForwardServiceServer struct {
	forwardv1.UnimplementedForwardServiceServer

	Svc *srv.Service
}

func NewForwardServiceServer(ctx *srv.Service) *ForwardServiceServer {
	return &ForwardServiceServer{
		Svc: ctx,
	}
}

func (s *ForwardServiceServer) PushForward(ctx context.Context, in *forwardv1.PushForwardRequest) (
	*forwardv1.PushForwardResponse, error) {

	if len(in.Targets) == 0 {
		return &forwardv1.PushForwardResponse{}, nil
	}

	reqs := make([]*srv.ForwardReq, 0, len(in.Targets))
	for _, t := range in.GetTargets() {
		reqs = append(reqs, &srv.ForwardReq{
			SessId:     t.Id,
			Data:       t.Data,
			ForwardCnt: t.ForwardCnt,
		})
	}

	err := s.Svc.ForwardService.Forward(ctx, reqs)
	if err != nil {
		return nil, err
	}

	return &forwardv1.PushForwardResponse{}, nil
}
