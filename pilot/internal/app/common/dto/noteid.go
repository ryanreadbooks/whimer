package dto

import (
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notecreator/errors"
	notevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
)

type NoteIdReq struct {
	NoteId notevo.NoteId `json:"note_id" path:"note_id" form:"note_id"`
}

func (r *NoteIdReq) Validate() error {
	if r == nil {
		return errors.ErrNilArg
	}

	if r.NoteId <= 0 {
		return errors.ErrNoteNotFound
	}

	return nil
}
