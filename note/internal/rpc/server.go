package rpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/global"
	mgtp "github.com/ryanreadbooks/whimer/note/internal/model/note"
	"github.com/ryanreadbooks/whimer/note/internal/svc"
	sdk "github.com/ryanreadbooks/whimer/note/sdk/v1"
)

type NoteServer struct {
	sdk.UnimplementedNoteServer

	Svc *svc.ServiceContext
}

func NewNoteServer(svc *svc.ServiceContext) *NoteServer {
	return &NoteServer{
		Svc: svc,
	}
}

func (s *NoteServer) IsUserOwnNote(ctx context.Context, in *sdk.IsUserOwnNoteReq) (*sdk.IsUserOwnNoteRes, error) {
	nid := s.Svc.NoteSvc.NoteIdConfuser.ConfuseU(in.NoteId)
	owner, err := s.Svc.NoteSvc.GetNoteOwner(ctx, nid)
	if err != nil {
		return nil, err
	}

	return &sdk.IsUserOwnNoteRes{Uid: in.Uid, Result: owner == in.Uid}, nil
}

// 判断笔记是否存在
func (s *NoteServer) IsNoteExist(ctx context.Context, in *sdk.IsNoteExistReq) (*sdk.IsNoteExistRes, error) {
	nid := s.Svc.NoteSvc.NoteIdConfuser.ConfuseU(in.NoteId)
	ok, err := s.Svc.NoteSvc.IsNoteExist(ctx, nid)
	if err != nil {
		return nil, err
	}

	return &sdk.IsNoteExistRes{Exist: ok}, nil
}

// 创建笔记
func (s *NoteServer) CreateNote(ctx context.Context, in *sdk.CreateNoteReq) (*sdk.CreateNoteRes, error) {
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

	return &sdk.CreateNoteRes{NoteId: noteId}, nil
}

// 更新笔记
func (s *NoteServer) UpdateNote(ctx context.Context, in *sdk.UpdateNoteReq) (*sdk.UpdateNoteRes, error) {
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

	return &sdk.UpdateNoteRes{NoteId: req.NoteId}, nil
}

// 删除笔记
func (s *NoteServer) DeleteNote(ctx context.Context, in *sdk.DeleteNoteReq) (*sdk.DeleteNoteRes, error) {
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

	return &sdk.DeleteNoteRes{}, nil
}

// 获取笔记的信息
func (s *NoteServer) GetNote(ctx context.Context, in *sdk.GetNoteReq) (*sdk.NoteItem, error) {
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
func (s *NoteServer) ListNote(ctx context.Context, in *sdk.ListNoteReq) (*sdk.ListNoteRes, error) {
	data, err := s.Svc.NoteSvc.List(ctx)
	if err != nil {
		return nil, err
	}

	items := make([]*sdk.NoteItem, 0, len(data.Items))
	for _, item := range data.Items {
		items = append(items, item.AsPb())
	}

	return &sdk.ListNoteRes{Items: items}, nil
}

func (s *NoteServer) GetUploadAuth(ctx context.Context, in *sdk.GetUploadAuthReq) (*sdk.GetUploadAuthRes, error) {
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
