package rpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/global"
	mgtp "github.com/ryanreadbooks/whimer/note/internal/model/note"
	"github.com/ryanreadbooks/whimer/note/internal/svc"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
)

var (
	NoteServiceName = notev1.NoteAdminService_ServiceDesc.ServiceName
)

type NoteAdminServiceServer struct {
	notev1.UnimplementedNoteAdminServiceServer

	Svc *svc.ServiceContext
}

func NewNoteAdminServiceServer(svc *svc.ServiceContext) *NoteAdminServiceServer {
	return &NoteAdminServiceServer{
		Svc: svc,
	}
}

func (s *NoteAdminServiceServer) IsUserOwnNote(ctx context.Context, in *notev1.IsUserOwnNoteRequest) (
	*notev1.IsUserOwnNoteResponse, error) {
	owner, err := s.Svc.NoteAdminSvc.GetNoteOwner(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.IsUserOwnNoteResponse{Uid: in.Uid, Result: owner == in.Uid}, nil
}

// 判断笔记是否存在
func (s *NoteAdminServiceServer) IsNoteExist(ctx context.Context, in *notev1.IsNoteExistRequest) (
	*notev1.IsNoteExistResponse, error) {
	ok, err := svc.IsNoteExist(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.IsNoteExistResponse{Exist: ok}, nil
}

// 创建笔记
func (s *NoteAdminServiceServer) CreateNote(ctx context.Context, in *notev1.CreateNoteRequest) (
	*notev1.CreateNoteResponse, error) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	images := make([]mgtp.CreateReqImage, 0, len(in.Images))
	for _, img := range in.Images {
		images = append(images, mgtp.CreateReqImage{
			FileId: img.FileId,
		})
	}

	var req = mgtp.CreateReq{
		Basic: mgtp.CreateReqBasic{
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
	noteId, err := s.Svc.NoteAdminSvc.Create(ctx, &req)
	if err != nil {
		return nil, err
	}

	return &notev1.CreateNoteResponse{NoteId: noteId}, nil
}

// 更新笔记
func (s *NoteAdminServiceServer) UpdateNote(ctx context.Context, in *notev1.UpdateNoteRequest) (
	*notev1.UpdateNoteResponse, error) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	images := make([]mgtp.CreateReqImage, 0, len(in.Note.Images))
	for _, img := range in.Note.Images {
		images = append(images, mgtp.CreateReqImage{
			FileId: img.FileId,
		})
	}

	var req = mgtp.UpdateReq{
		NoteId: in.NoteId,
		CreateReq: mgtp.CreateReq{
			Basic: mgtp.CreateReqBasic{
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

	err := s.Svc.NoteAdminSvc.Update(ctx, &req)
	if err != nil {
		return nil, err
	}

	return &notev1.UpdateNoteResponse{NoteId: req.NoteId}, nil
}

// 删除笔记
func (s *NoteAdminServiceServer) DeleteNote(ctx context.Context, in *notev1.DeleteNoteRequest) (
	*notev1.DeleteNoteResponse, error) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	var req = mgtp.DeleteReq{
		NoteId: in.NoteId,
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	err := s.Svc.NoteAdminSvc.Delete(ctx, &req)
	if err != nil {
		return nil, err
	}

	return &notev1.DeleteNoteResponse{}, nil
}

// 获取笔记的信息
func (s *NoteAdminServiceServer) GetNote(ctx context.Context, in *notev1.GetNoteRequest) (
	*notev1.GetNoteResponse, error) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	if in.NoteId == 0 {
		return nil, global.ErrNoteNotFound
	}

	data, err := s.Svc.NoteAdminSvc.GetNote(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.GetNoteResponse{Note: data.AsPb()}, nil
}

// 列出笔记
func (s *NoteAdminServiceServer) ListNote(ctx context.Context, in *notev1.ListNoteRequest) (
	*notev1.ListNoteResponse, error) {
	data, err := s.Svc.NoteAdminSvc.List(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*notev1.NoteItem, 0, len(data.Items))
	for _, item := range data.Items {
		items = append(items, item.AsPb())
	}

	return &notev1.ListNoteResponse{Items: items}, nil
}

func (s *NoteAdminServiceServer) GetUploadAuth(ctx context.Context, in *notev1.GetUploadAuthRequest) (
	*notev1.GetUploadAuthResponse, error) {
	var req = mgtp.UploadAuthReq{
		Resource: in.Resource,
		Source:   in.Source,
		MimeType: in.MimeType,
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	data, err := s.Svc.NoteAdminSvc.UploadAuth(ctx, &req)
	if err != nil {
		return nil, err
	}

	return data.AsPb(), nil
}
