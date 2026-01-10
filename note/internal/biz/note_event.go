package biz

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/data"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	eventmodel "github.com/ryanreadbooks/whimer/note/internal/model/event"
)

type NoteEventBiz struct {
	data *data.Data
}

func NewNoteEventBiz(dt *data.Data) *NoteEventBiz {
	return &NoteEventBiz{data: dt}
}

func (b *NoteEventBiz) NotePublished(ctx context.Context, note *model.Note) error {
	return b.data.NoteEventBus.NotePublished(ctx, note)
}

func (b *NoteEventBiz) NoteDeleted(ctx context.Context, note *model.Note, reason eventmodel.NoteDeleteReason) error {
	return b.data.NoteEventBus.NoteDeleted(ctx, note, reason)
}

func (b *NoteEventBiz) NoteLiked(ctx context.Context, noteId, userId, ownerId int64, isLiked bool) error {
	return b.data.NoteEventBus.NoteLiked(ctx, noteId, userId, ownerId, isLiked)
}
