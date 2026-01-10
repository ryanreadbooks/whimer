package biz

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/data"
)

type Biz struct {
	data *data.Data

	Note      *NoteBiz
	Interact  *NoteInteractBiz
	Creator   *NoteCreatorBiz
	Procedure *NoteProcedureBiz
	NoteEvent *NoteEventBiz
}

func New(dt *data.Data) *Biz {
	note := NewNoteBiz(dt)
	creator := NewNoteCreatorBiz(dt, note)
	interact := NewNoteInteractBiz(dt, note)
	procedure := NewNoteProcedureBiz(dt)
	noteEvent := NewNoteEventBiz(dt)

	return &Biz{
		data:      dt,
		Note:      note,
		Interact:  interact,
		Creator:   creator,
		Procedure: procedure,
		NoteEvent: noteEvent,
	}
}

func (b *Biz) Tx(ctx context.Context, fn func(ctx context.Context) error) error {
	return b.data.DB().Transact(ctx, fn)
}

func (b *Biz) Data() *data.Data {
	return b.data
}
