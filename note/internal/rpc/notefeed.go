package rpc

import (
	"github.com/ryanreadbooks/whimer/note/internal/svc"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
)

var (
	NoteFeedServiceName = notev1.NoteFeedService_ServiceDesc.ServiceName
)

type NoteFeedServiceServer struct {
	notev1.UnimplementedNoteFeedServiceServer

	Svc *svc.ServiceContext
}

func NewNoteFeedServiceServer(svc *svc.ServiceContext) *NoteFeedServiceServer {
	return &NoteFeedServiceServer{
		Svc: svc,
	}
}
