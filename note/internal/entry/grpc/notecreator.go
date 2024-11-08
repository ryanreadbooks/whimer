package grpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	"github.com/ryanreadbooks/whimer/note/internal/srv"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
)

type NoteCreatorServiceServer struct {
	notev1.UnimplementedNoteCreatorServiceServer

	Srv *srv.ServiceContext
}

func NewNoteAdminServiceServer(srv *srv.ServiceContext) *NoteCreatorServiceServer {
	return &NoteCreatorServiceServer{
		Srv: srv,
	}
}

func (s *NoteCreatorServiceServer) IsUserOwnNote(ctx context.Context, in *notev1.IsUserOwnNoteRequest) (
	*notev1.IsUserOwnNoteResponse, error) {
	owner, err := s.Srv.NoteCreatorSrv.GetNoteOwner(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.IsUserOwnNoteResponse{Uid: in.Uid, Result: owner == in.Uid}, nil
}

// 判断笔记是否存在
func (s *NoteCreatorServiceServer) IsNoteExist(ctx context.Context, in *notev1.IsNoteExistRequest) (
	*notev1.IsNoteExistResponse, error) {
	ok, err := s.Srv.NoteCreatorSrv.IsNoteExist(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.IsNoteExistResponse{Exist: ok}, nil
}

// 创建笔记
func (s *NoteCreatorServiceServer) CreateNote(ctx context.Context, in *notev1.CreateNoteRequest) (
	*notev1.CreateNoteResponse, error) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	images := make([]model.CreateNoteRequestImage, 0, len(in.Images))
	for _, img := range in.Images {
		images = append(images, model.CreateNoteRequestImage{
			FileId: img.FileId,
		})
	}

	var req = model.CreateNoteRequest{
		Basic: model.CreateNoteRequestBasic{
			Title:   in.Basic.Title,
			Desc:    in.Basic.Desc,
			Privacy: int(in.Basic.Privacy),
		},
		Images: images,
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// service to create note
	noteId, err := s.Srv.NoteCreatorSrv.Create(ctx, &req)
	if err != nil {
		return nil, err
	}

	return &notev1.CreateNoteResponse{NoteId: noteId}, nil
}

// 更新笔记
func (s *NoteCreatorServiceServer) UpdateNote(ctx context.Context, in *notev1.UpdateNoteRequest) (
	*notev1.UpdateNoteResponse, error) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	images := make([]model.CreateNoteRequestImage, 0, len(in.Note.Images))
	for _, img := range in.Note.Images {
		images = append(images, model.CreateNoteRequestImage{
			FileId: img.FileId,
		})
	}

	var req = model.UpdateNoteRequest{
		NoteId: in.NoteId,
		CreateNoteRequest: model.CreateNoteRequest{
			Basic: model.CreateNoteRequestBasic{
				Title:   in.Note.Basic.Title,
				Desc:    in.Note.Basic.Desc,
				Privacy: int(in.Note.Basic.Privacy),
			},
			Images: images,
		},
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	err := s.Srv.NoteCreatorSrv.Update(ctx, &req)
	if err != nil {
		return nil, err
	}

	return &notev1.UpdateNoteResponse{NoteId: req.NoteId}, nil
}

// 删除笔记
func (s *NoteCreatorServiceServer) DeleteNote(ctx context.Context, in *notev1.DeleteNoteRequest) (
	*notev1.DeleteNoteResponse, error) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	var req = model.DeleteNoteRequest{
		NoteId: in.NoteId,
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	err := s.Srv.NoteCreatorSrv.Delete(ctx, &req)
	if err != nil {
		return nil, err
	}

	return &notev1.DeleteNoteResponse{}, nil
}

// 用于笔记作者获取笔记的详细信息
func (s *NoteCreatorServiceServer) GetNote(ctx context.Context, in *notev1.GetNoteRequest) (
	*notev1.GetNoteResponse, error) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	if in.NoteId == 0 {
		return nil, global.ErrNoteNotFound
	}

	data, err := s.Srv.NoteCreatorSrv.GetNote(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.GetNoteResponse{Note: data.AsPb()}, nil
}

// 列出笔记
func (s *NoteCreatorServiceServer) ListNote(ctx context.Context, in *notev1.ListNoteRequest) (
	*notev1.ListNoteResponse, error) {
	data, err := s.Srv.NoteCreatorSrv.List(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*notev1.NoteItem, 0, len(data.Items))
	for _, item := range data.Items {
		items = append(items, item.AsPb())
	}

	return &notev1.ListNoteResponse{Items: items}, nil
}

func (s *NoteCreatorServiceServer) GetUploadAuth(ctx context.Context, in *notev1.GetUploadAuthRequest) (
	*notev1.GetUploadAuthResponse, error) {
	var req = model.UploadAuthRequest{
		Resource: in.Resource,
		Source:   in.Source,
		MimeType: in.MimeType,
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	data, err := s.Srv.NoteCreatorSrv.UploadAuth(ctx, &req)
	if err != nil {
		return nil, err
	}

	return data.AsPb(), nil
}
