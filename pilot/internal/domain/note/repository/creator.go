package repository

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/entity"
)

// 远程依赖
type NoteCreatorAdapter interface {
	GetNote(ctx context.Context, noteId int64) (*entity.CreatorNote, error)
	CreateNote(ctx context.Context, params *entity.CreateNoteParams) (int64, error)
	UpdateNote(ctx context.Context, params *entity.UpdateNoteParams) (int64, error)
	DeleteNote(ctx context.Context, noteId int64) error
	PageListNotes(ctx context.Context, params *entity.PageListNotesParams) (*entity.PageListNotesResult, error)

	// Tag 相关
	AddTag(ctx context.Context, name string) (int64, error)
	SearchTags(ctx context.Context, name string) ([]*entity.SearchedNoteTag, error)
	GetTag(ctx context.Context, tagId int64) (*entity.NoteTag, error)

	// 用户投稿数（包含私密）
	GetPostedCount(ctx context.Context, uid int64) (int64, error)
}
