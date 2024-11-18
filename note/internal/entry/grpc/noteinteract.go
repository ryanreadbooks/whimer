package grpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/srv"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
)

type NoteInteractServiceServer struct {
	notev1.UnimplementedNoteInteractServiceServer

	Srv *srv.Service
}

func NewNoteInteractServiceServer(svc *srv.Service) *NoteInteractServiceServer {
	return &NoteInteractServiceServer{
		Srv: svc,
	}
}

// 点赞笔记
func (s *NoteInteractServiceServer) LikeNote(ctx context.Context, in *notev1.LikeNoteRequest) (
	*notev1.LikeNoteResponse, error) {
	return s.Srv.NoteInteractSrv.LikeNote(ctx, in)
}

// 获取笔记点赞数量
func (s *NoteInteractServiceServer) GetNoteLikes(ctx context.Context, in *notev1.GetNoteLikesRequest) (
	*notev1.GetNoteLikesResponse, error) {
	likes, err := s.Srv.NoteInteractSrv.GetNoteLikes(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.GetNoteLikesResponse{Likes: likes, NoteId: in.NoteId}, nil
}

// 检查用户点赞状态
func (s *NoteInteractServiceServer) CheckUserLikeStatus(ctx context.Context, in *notev1.CheckUserLikeStatusRequest) (
	*notev1.CheckUserLikeStatusResponse, error) {
	resp, err := s.Srv.NoteInteractSrv.CheckUserLikeStatus(ctx, in.Uid, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.CheckUserLikeStatusResponse{Liked: resp}, nil
}

// 获取用户和某笔记的交互状态，包括是否点赞等
func (s *NoteInteractServiceServer) GetNoteInteraction(ctx context.Context, in *notev1.GetNoteInteractionRequest) (
	*notev1.GetNoteInteractionResponse, error) {
	resp, err := s.Srv.NoteInteractSrv.GetNoteInteraction(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.GetNoteInteractionResponse{
		Interaction: &notev1.NoteInteraction{
			Liked: resp.Liked,
		},
	}, nil
}
