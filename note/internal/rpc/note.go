package rpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/global"
	mgtp "github.com/ryanreadbooks/whimer/note/internal/model/note"
	"github.com/ryanreadbooks/whimer/note/internal/svc"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
)

var (
	NoteServiceName = notev1.NoteService_ServiceDesc.ServiceName
)

type NoteServiceServer struct {
	notev1.UnimplementedNoteServiceServer

	Svc *svc.ServiceContext
}

func NewNoteServiceServer(svc *svc.ServiceContext) *NoteServiceServer {
	return &NoteServiceServer{
		Svc: svc,
	}
}

func (s *NoteServiceServer) IsUserOwnNote(ctx context.Context, in *notev1.IsUserOwnNoteReq) (*notev1.IsUserOwnNoteRes, error) {
	owner, err := s.Svc.NoteSvc.GetNoteOwner(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.IsUserOwnNoteRes{Uid: in.Uid, Result: owner == in.Uid}, nil
}

// 判断笔记是否存在
func (s *NoteServiceServer) IsNoteExist(ctx context.Context, in *notev1.IsNoteExistReq) (*notev1.IsNoteExistRes, error) {
	ok, err := s.Svc.NoteSvc.IsNoteExist(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.IsNoteExistRes{Exist: ok}, nil
}

// 创建笔记
func (s *NoteServiceServer) CreateNote(ctx context.Context, in *notev1.CreateNoteReq) (*notev1.CreateNoteRes, error) {
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
	noteId, err := s.Svc.NoteSvc.Create(ctx, &req)
	if err != nil {
		return nil, err
	}

	return &notev1.CreateNoteRes{NoteId: noteId}, nil
}

// 更新笔记
func (s *NoteServiceServer) UpdateNote(ctx context.Context, in *notev1.UpdateNoteReq) (*notev1.UpdateNoteRes, error) {
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

	err := s.Svc.NoteSvc.Update(ctx, &req)
	if err != nil {
		return nil, err
	}

	return &notev1.UpdateNoteRes{NoteId: req.NoteId}, nil
}

// 删除笔记
func (s *NoteServiceServer) DeleteNote(ctx context.Context, in *notev1.DeleteNoteReq) (*notev1.DeleteNoteRes, error) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	var req = mgtp.DeleteReq{
		NoteId: in.NoteId,
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	err := s.Svc.NoteSvc.Delete(ctx, &req)
	if err != nil {
		return nil, err
	}

	return &notev1.DeleteNoteRes{}, nil
}

// 获取笔记的信息
func (s *NoteServiceServer) GetNote(ctx context.Context, in *notev1.GetNoteReq) (*notev1.NoteItem, error) {
	if in == nil {
		return nil, global.ErrNilReq
	}

	if in.NoteId == 0 {
		return nil, global.ErrNoteNotFound
	}

	data, err := s.Svc.NoteSvc.GetNote(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return data.AsPb(), nil
}

// 列出笔记
func (s *NoteServiceServer) ListNote(ctx context.Context, in *notev1.ListNoteReq) (*notev1.ListNoteRes, error) {
	data, err := s.Svc.NoteSvc.List(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*notev1.NoteItem, 0, len(data.Items))
	for _, item := range data.Items {
		items = append(items, item.AsPb())
	}

	return &notev1.ListNoteRes{Items: items}, nil
}

func (s *NoteServiceServer) GetUploadAuth(ctx context.Context, in *notev1.GetUploadAuthReq) (*notev1.GetUploadAuthRes, error) {
	var req = mgtp.UploadAuthReq{
		Resource: in.Resource,
		Source:   in.Source,
		MimeType: in.MimeType,
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	data, err := s.Svc.NoteSvc.UploadAuth(ctx, &req)
	if err != nil {
		return nil, err
	}

	return data.AsPb(), nil
}

func (s *NoteServiceServer) LikeNote(ctx context.Context, in *notev1.LikeNoteReq) (*notev1.LikeNoteRes, error) {
	return s.Svc.NoteSvc.LikeNote(ctx, in)
}

func (s *NoteServiceServer) GetNoteLikes(ctx context.Context, in *notev1.GetNoteLikesReq) (
	*notev1.GetNoteLikesRes, error) {
	likes, err := s.Svc.NoteSvc.GetNoteLikes(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.GetNoteLikesRes{Likes: likes}, nil
}
