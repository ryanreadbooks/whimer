package rpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/svc"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"
)

type NoteInteractServiceServer struct {
	notev1.UnimplementedNoteInteractServiceServer

	Svc *svc.ServiceContext
}

func NewNoteInteractServiceServer(svc *svc.ServiceContext) *NoteInteractServiceServer {
	return &NoteInteractServiceServer{
		Svc: svc,
	}
}

// 点赞笔记
func (s *NoteInteractServiceServer) LikeNote(ctx context.Context, in *notev1.LikeNoteRequest) (
	*notev1.LikeNoteResponse, error) {
	return s.Svc.NoteInteractSvc.LikeNote(ctx, in)
}

// 获取笔记点赞数量
func (s *NoteInteractServiceServer) GetNoteLikes(ctx context.Context, in *notev1.GetNoteLikesRequest) (
	*notev1.GetNoteLikesResponse, error) {
	likes, err := s.Svc.NoteInteractSvc.GetNoteLikes(ctx, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.GetNoteLikesResponse{Likes: likes}, nil
}

// 检查用户点赞状态
func (s *NoteInteractServiceServer) CheckUserLikeStatus(ctx context.Context, in *notev1.CheckUserLikeStatusRequest) (
	*notev1.CheckUserLikeStatusResponse, error) {
	resp, err := s.Svc.NoteInteractSvc.CheckUserLikeStatus(ctx, in.Uid, in.NoteId)
	if err != nil {
		return nil, err
	}

	return &notev1.CheckUserLikeStatusResponse{Liked: resp}, nil
}
