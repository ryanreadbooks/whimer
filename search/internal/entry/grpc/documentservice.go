package grpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
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

	for _, t := range in.GetNoteTags() {
		if len(t.Id) == 0 {
			return nil, xerror.ErrArgs.Msg("note tags contain empty tag id")
		}
	}

	err := s.svc.DocumentSrv.AddNoteTagDocs(ctx, in.GetNoteTags())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// 批量添加笔记文档
func (s *DocumentServiceServerImpl) BatchAddNote(ctx context.Context,
	in *searchv1.BatchAddNoteRequest) (*searchv1.BatchAddNoteResponse, error) {

	var resp = &searchv1.BatchAddNoteResponse{}
	if len(in.GetNotes()) == 0 {
		return resp, nil
	}

	for _, n := range in.GetNotes() {
		if len(n.NoteId) == 0 {
			return nil, xerror.ErrArgs.Msg("notes contain empty note id")
		}
	}

	err := s.svc.DocumentSrv.AddNoteDocs(ctx, in.GetNotes())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *DocumentServiceServerImpl) BatchDeleteNote(ctx context.Context,
	in *searchv1.BatchDeleteNoteRequest) (*searchv1.BatchDeleteNoteResponse, error) {

	var resp = &searchv1.BatchDeleteNoteResponse{}
	if len(in.GetIds()) == 0 {
		return resp, nil
	}

	for _, id := range in.GetIds() {
		if len(id) == 0 {
			return nil, xerror.ErrArgs.Msg("note ids contain empty note id")
		}
	}

	err := s.svc.DocumentSrv.DeleteNoteDocs(ctx, in.GetIds())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// 更新笔记点赞数量
func (s *DocumentServiceServerImpl) BatchUpdateNoteLikeCount(ctx context.Context,
	in *searchv1.BatchUpdateNoteLikeCountRequest) (*searchv1.BatchUpdateNoteLikeCountResponse, error) {
	var resp = &searchv1.BatchUpdateNoteLikeCountResponse{}
	if len(in.GetCounts()) == 0 {
		return resp, nil
	}

	err := s.svc.DocumentSrv.UpdateNoteDocLikeCount(ctx, in.GetCounts())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// 更新笔记评论数量
func (s *DocumentServiceServerImpl) BatchUpdateNoteCommentCount(ctx context.Context,
	in *searchv1.BatchUpdateNoteCommentCountRequest) (*searchv1.BatchUpdateNoteCommentCountResponse, error) {
	var resp = &searchv1.BatchUpdateNoteCommentCountResponse{}
	if len(in.GetCounts()) == 0 {
		return resp, nil
	}

	err := s.svc.DocumentSrv.UpdateNoteDocCommentCount(ctx, in.GetCounts())
	if err != nil {
		return nil, err
	}

	return resp, nil
}
