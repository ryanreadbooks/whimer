package grpc

import (
	"context"

	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/note/internal/srv"
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

// 批量检查用户点赞状态
func (s *NoteInteractServiceServer) BatchCheckUserLikeStatus(ctx context.Context,
	in *notev1.BatchCheckUserLikeStatusRequest) (
	*notev1.BatchCheckUserLikeStatusResponse, error) {

	var uidNoteIds = make(map[int64][]int64, len(in.Mappings))
	for uid, m := range in.GetMappings() {
		uidNoteIds[uid] = append(uidNoteIds[uid], m.NoteIds...)
	}

	resp, err := s.Srv.NoteInteractSrv.BatchCheckUserLikeStatus(ctx, uidNoteIds)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*notev1.LikeStatusList)
	for uid, status := range resp {
		list := make([]*notev1.LikeStatus, 0, len(status))
		for _, s := range status {
			list = append(list, &notev1.LikeStatus{
				NoteId: s.NoteId,
				Liked:  s.Liked,
			})
		}
		result[uid] = &notev1.LikeStatusList{
			List: list,
		}
	}

	return &notev1.BatchCheckUserLikeStatusResponse{Results: result}, nil
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

// 列出用户点赞过的笔记
func (s *NoteInteractServiceServer) PageListUserLikedNote(ctx context.Context, in *notev1.PageListUserLikedNoteRequest) (
	*notev1.PageListUserLikedNoteResponse, error) {

	var resp = notev1.PageListUserLikedNoteResponse{}
	if in.Uid == 0 {
		return &resp, nil
	}

	targetNoteIds, pgres, err := s.Srv.NoteInteractSrv.PageListUserLikedNoteIds(ctx, in)
	if err != nil {
		return nil, err
	}

	if len(targetNoteIds) == 0 {
		return &resp, nil
	}

	targetNotes, err := s.Srv.NoteFeedSrv.BatchGetNoteDetail(ctx, targetNoteIds)
	if err != nil {
		return nil, err
	}

	items := make([]*notev1.FeedNoteItem, 0, len(targetNotes))
	// return ids order should be retain
	for _, noteId := range targetNoteIds {
		if targetNote, ok := targetNotes[noteId]; ok && targetNote != nil {
			items = append(items, targetNote.AsFeedPb())
		}
	}
	resp.NextCursor = pgres.NextCursor
	resp.HasNext = pgres.HasNext
	resp.Items = items

	return &resp, nil
}
