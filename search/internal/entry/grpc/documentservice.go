package grpc

import (
	"context"

	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
	"github.com/ryanreadbooks/whimer/search/internal/srv"
)

type DocumentServiceServerImpl struct {
	searchv1.UnimplementedDocumentServiceServer

	svc *srv.Service
}

func NewDocumentService(svc *srv.Service) searchv1.DocumentServiceServer {
	return &DocumentServiceServerImpl{
		svc: svc,
	}
}

// 添加笔记标签文档
func (s *DocumentServiceServerImpl) BatchAddNoteTag(ctx context.Context,
	in *searchv1.BatchAddNoteTagRequest) (*searchv1.BatchAddNoteTagResponse, error) {

	var resp = &searchv1.BatchAddNoteTagResponse{}
	if len(in.GetNoteTags()) == 0 {
		return resp, nil
	}

	err := s.svc.DocumentSrv.AddNoteTagDocs(ctx, in.GetNoteTags())
	if err != nil {
		return nil, err
	}

	return resp, nil
}
