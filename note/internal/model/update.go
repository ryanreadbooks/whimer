package model

import (
	"github.com/ryanreadbooks/whimer/note/internal/global"
)

type UpdateNoteRequest struct {
	NoteId int64 `json:"note_id"`
	CreateNoteRequest
}

func (r *UpdateNoteRequest) Validate() error {
	if r == nil {
		return global.ErrNilReq
	}

	if r.NoteId == 0 {
		return global.ErrArgs.Msg("笔记不存在")
	}

	return r.CreateNoteRequest.Validate()
}
