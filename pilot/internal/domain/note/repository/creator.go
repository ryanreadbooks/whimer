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
}
