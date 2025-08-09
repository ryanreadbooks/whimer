package grpc

import (
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
