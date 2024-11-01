package rpc

import (
	"context"

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

func (s *NoteFeedServiceServer) RandomGet(ctx context.Context, in *notev1.RandomGetRequest) (
	*notev1.RandomGetResponse, error,
) {
	resp, err := s.Svc.NoteFeedSvc.FeedRandomGet(ctx, in.Count)
	if err != nil {
		return nil, err
	}

	items := make([]*notev1.NoteItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		items = append(items, item.AsPb())
	}

	return &notev1.RandomGetResponse{
		Items: items,
		Count: int32(len(items)),
	}, nil
}