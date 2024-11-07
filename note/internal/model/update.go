package model

import (
	"github.com/ryanreadbooks/whimer/note/internal/global"
)

type UpdateNoteRequest struct {
	NoteId uint64 `json:"note_id"`
	CreateNoteRequest
}

func (r *UpdateNoteRequest) Validate() error {
	if r == nil {
		return global.ErrNilReq
	}

	if r.NoteId == 0 {
		return global.ErrArgs.Msg("笔记id错误")
	}

	return r.CreateNoteRequest.Validate()
}

type UpdateResponse struct {
	NoteId string `json:"note_id"`
}
