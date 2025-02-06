package grpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	"github.com/ryanreadbooks/whimer/note/internal/srv"
)

type NoteCreatorServiceServer struct {
	notev1.UnimplementedNoteCreatorServiceServer

	Srv *srv.Service
}

func NewNoteAdminServiceServer(srv *srv.Service) *NoteCreatorServiceServer {
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
			Width:  img.Width,
			Height: img.Height,
			Format: img.Format,
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
	data, nextPage, err := s.Srv.NoteCreatorSrv.List(ctx, in.Cursor, in.Count)
	if err != nil {
		return nil, err
	}

	items := make([]*notev1.NoteItem, 0, len(data.Items))
	for _, item := range data.Items {
		items = append(items, item.AsPb())
	}

	return &notev1.ListNoteResponse{
		Items:      items,
		NextCursor: nextPage.NextCursor,
		HasNext:    nextPage.HasNext}, nil
}

func (s *NoteCreatorServiceServer) GetUploadAuth(ctx context.Context, in *notev1.GetUploadAuthRequest) (
	*notev1.GetUploadAuthResponse, error) {
	return nil, xerror.ErrApiWentOffline

	// var req = model.UploadAuthRequest{
	// 	Resource: in.Resource,
	// 	Source:   in.Source,
	// }

	// if err := req.Validate(); err != nil {
	// 	return nil, err
	// }

	// data, err := s.Srv.NoteCreatorSrv.UploadAuth(ctx, &req)
	// if err != nil {
	// 	return nil, err
	// }

	// return data.AsPb(), nil
}

// Deprecated
//
// 批量获取上传凭证
func (s *NoteCreatorServiceServer) BatchGetUploadAuth(ctx context.Context,
	in *notev1.BatchGetUploadAuthRequest) (
	*notev1.BatchGetUploadAuthResponse, error,
) {

	return nil, xerror.ErrApiWentOffline

	// var req = model.UploadAuthRequest{
	// 	Resource: in.Resource,
	// 	Source:   in.Source,
	// 	Count:    in.Count,
	// }

	// if err := req.Validate(); err != nil {
	// 	return nil, err
	// }

	// data, err := s.Srv.NoteCreatorSrv.BatchGetUploadAuth(ctx, &req)
	// if err != nil {
	// 	return nil, err
	// }

	// resp := notev1.BatchGetUploadAuthResponse{}
	// for _, d := range data {
	// 	resp.Tickets = append(resp.Tickets, d.AsPb())
	// }

	// return &resp, nil
}

func (s *NoteCreatorServiceServer) GetPostedCount(ctx context.Context, in *notev1.GetPostedCountRequest) (
	*notev1.GetPostedCountResponse, error) {
	cnt, err := s.Srv.NoteCreatorSrv.GetPostedCount(ctx, in.Uid)
	if err != nil {
		return nil, err
	}

	return &notev1.GetPostedCountResponse{Count: cnt}, nil
}

func (s *NoteCreatorServiceServer) BatchGetUploadAuthV2(ctx context.Context,
	in *notev1.BatchGetUploadAuthV2Request) (*notev1.BatchGetUploadAuthV2Response, error) {

	var req = model.UploadAuthRequest{
		Resource: in.Resource,
		Source:   in.Source,
		Count:    in.Count,
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	res, err := s.Srv.NoteCreatorSrv.BatchGetUploadAuthSTS(ctx, &req)
	if err != nil {
		return nil, err
	}

	return res.AsPb(), nil
}
