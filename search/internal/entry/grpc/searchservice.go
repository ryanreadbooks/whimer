package grpc

import (
	"context"
	"strings"

	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
	"github.com/ryanreadbooks/whimer/search/internal/srv"
)

type SearchServiceServerImpl struct {
	searchv1.UnimplementedSearchServiceServer

	svc *srv.Service
}

func NewSearchService(svc *srv.Service) searchv1.SearchServiceServer {
	return &SearchServiceServerImpl{
		svc: svc,
	}
}

func (s *SearchServiceServerImpl) SearchNoteTags(ctx context.Context, in *searchv1.SearchNoteTagsRequest) (
	*searchv1.SearchNoteTagsResponse, error) {

	var resp = &searchv1.SearchNoteTagsResponse{}
	in.Text = strings.TrimSpace(in.Text)
	if len(in.Text) == 0 {
		return resp, nil
	}

	if in.Page <= 0 {
		in.Page = 1
	}

	if in.Count >= 30 || in.Count <= 0 {
		in.Count = 30
	}

	items, total, err := s.svc.SearchSrv.SearchNoteTags(ctx, in.Text, in.Page, in.Count)
	if err != nil {
		return nil, err
	}
	resp.Total = total
	for _, item := range items {
		resp.Items = append(resp.Items, &searchv1.NoteTag{
			Id:    item.Id,
			Name:  item.Name,
			Ctime: item.Ctime,
		})
	}

	return resp, nil
}
