package repository

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"
)

type CursorPageResult struct {
	HasNext    bool
	NextCursor int64
}

type CursorPageResultV2 struct {
	HasNext    bool
	NextCursor string
}

type NoteFeedAdapter interface {
	// 获取笔记
	RandomGet(ctx context.Context, count int32) ([]*entity.FeedNote, error)

	GetNote(ctx context.Context, noteId int64) (*entity.FeedNote, *entity.FeedNoteExt, error)

	BatchGetNotes(ctx context.Context, noteIds []int64) (map[int64]*entity.FeedNote, error)

	// 获取用户的笔记
	ListUserNote(ctx context.Context, uid int64, cursor int64, count int32) ([]*entity.FeedNote, *CursorPageResult, error)

	// 获取笔记作者uid
	GetNoteAuthorUid(ctx context.Context, noteId int64) (int64, error)

	// 获取用户点赞过的笔记
	ListUserLikedNote(ctx context.Context, uid int64, cursor string, count int32) ([]*entity.FeedNote, *CursorPageResultV2, error)

	// 用户公开投稿数
	GetPublicPostedCount(ctx context.Context, uid int64) (int64, error)

	// 获取用户最近发布
	GetUserRecentPost(ctx context.Context, uid int64, count int32) ([]*entity.RecentPost, error)
}
