package rpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/svc"
	"github.com/ryanreadbooks/whimer/note/sdk"
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

// 获取笔记的信息
func (s *NoteServer) GetNote(ctx context.Context, in *sdk.GetNoteReq) (*sdk.GetNoteRes, error) {

	return &sdk.GetNoteRes{}, nil
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
