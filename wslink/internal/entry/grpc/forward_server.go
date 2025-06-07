package grpc

import (
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
