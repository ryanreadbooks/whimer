package biz

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/infra"
)

type Biz struct {
	Note      NoteBiz
	Interact  NoteInteractBiz
	Creator   NoteCreatorBiz
	Procedure NoteProcedureBiz
}

func New() Biz {
	note := NewNoteBiz()
	creator := NewNoteCreatorBiz()
	interact := NewNoteInteractBiz()
	procedure := NewNoteProcedureBiz()
	return Biz{
		Note:      note,
		Interact:  interact,
		Creator:   creator,
		Procedure: procedure,
	}
}

func (b *Biz) Tx(ctx context.Context, fn func(ctx context.Context) error) error {
	return infra.Dao().DB().Transact(ctx, fn)
}
