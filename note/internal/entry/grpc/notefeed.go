package grpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/note/internal/model"
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

	return &notev1.GetFeedNoteResponse{
		Item: resp.AsFeedPb(),
		Ext: &notev1.FeedNoteItemExt{
			Tags: model.NoteTagListAsPb(resp.Tags),
		}}, nil
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

// 列出用户公开的笔记
func (s *NoteFeedServiceServer) ListFeedByUid(ctx context.Context, in *notev1.ListFeedByUidRequest) (
	*notev1.ListFeedByUidResponse, error) {
	if in.Count > 20 {
		in.Count = 20
	}

	resp, page, err := s.Srv.NoteFeedSrv.ListUserPublicNotes(ctx, in.Uid, in.Cursor, in.Count)
	if err != nil {
		return nil, err
	}

	items := make([]*notev1.FeedNoteItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		items = append(items, item.AsFeedPb())
	}

	return &notev1.ListFeedByUidResponse{
		Items:      items,
		NextCursor: page.NextCursor,
		HasNext:    page.HasNext,
	}, nil
}

// 按照tag id获取标签信息
func (s *NoteFeedServiceServer) GetTagInfo(ctx context.Context,
	in *notev1.GetTagInfoRequest) (*notev1.GetTagInfoResponse, error) {

	if in.Id == 0 {
		return nil, xerror.ErrArgs.Msg("tag not found")
	}

	tag, err := s.Srv.NoteFeedSrv.GetTagInfo(ctx, in.Id)
	if err != nil {
		return nil, err
	}

	return &notev1.GetTagInfoResponse{
		Tag: tag.AsPb(),
	}, nil
}
