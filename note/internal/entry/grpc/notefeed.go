package grpc

import (
	"context"

	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/note/internal/srv"
)

var (
	NoteFeedServiceName = notev1.NoteFeedService_ServiceDesc.ServiceName
)

type NoteFeedServiceServer struct {
	notev1.UnimplementedNoteFeedServiceServer

	Srv *srv.Service
}

func NewNoteFeedServiceServer(svc *srv.Service) *NoteFeedServiceServer {
	return &NoteFeedServiceServer{
		Srv: svc,
	}
}

func (s *NoteFeedServiceServer) RandomGet(ctx context.Context, in *notev1.RandomGetRequest) (
	*notev1.RandomGetResponse, error,
) {
	resp, err := s.Srv.NoteFeedSrv.FeedRandomGet(ctx, in.Count)
	if err != nil {
		return nil, err
	}

	items := make([]*notev1.FeedNoteItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		items = append(items, item.AsFeedPb())
	}

	return &notev1.RandomGetResponse{
		Items: items,
		Count: int32(len(items)),
	}, nil
}

func (s *NoteFeedServiceServer) GetFeedNote(ctx context.Context, in *notev1.GetFeedNoteRequest) (
	*notev1.GetFeedNoteResponse, error) {
	resp, err := s.Srv.NoteFeedSrv.GetNoteDetail(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.GetFeedNoteResponse{Item: resp.AsFeedPb()}, nil
}

// 获取指定用户的最近的笔记内容
func (s *NoteFeedServiceServer) GetUserRecentPost(ctx context.Context, in *notev1.GetUserRecentPostRequest) (
	*notev1.GetUserRecentPostResponse, error) {
	if in.Count > 5 {
		in.Count = 5
	}

	resp, err := s.Srv.NoteFeedSrv.GetUserRecentNotes(ctx, in.Uid, in.Count)
	if err != nil {
		return nil, err
	}

	items := make([]*notev1.FeedNoteItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		items = append(items, item.AsFeedPb())
	}

	return &notev1.GetUserRecentPostResponse{Items: items}, nil
}
