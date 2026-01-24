package repository

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
)

type GetLikeStatusParams struct {
	Uid    int64
	NoteId int64
}

type GetLikeStatusResult struct {
	Liked bool
}

type BatchGetLikeStatusParams struct {
	Uid     int64
	NoteIds []int64
}

type BatchGetLikeStatusResult struct {
	// noteId -> liked
	Liked map[int64]bool
}

type LikeNoteParams struct {
	Uid    int64
	NoteId int64
	Action vo.LikeAction
}

type NoteLikesAdapter interface {
	// 判断用户是否点赞过笔记
	GetLikeStatus(ctx context.Context, p *GetLikeStatusParams) (*GetLikeStatusResult, error)

	// 判断用户是否点赞多多篇笔记
	BatchGetLikeStatus(ctx context.Context, p *BatchGetLikeStatusParams) (*BatchGetLikeStatusResult, error)

	// 点赞/取消点赞笔记
	LikeNote(ctx context.Context, p *LikeNoteParams) error

	// 获取笔记点赞数量
	GetLikeCount(ctx context.Context, noteId int64) (int64, error)
}
