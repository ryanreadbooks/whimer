package model

import "github.com/ryanreadbooks/whimer/note/internal/global"

type DeleteNoteRequest struct {
	NoteId int64 `json:"note_id"`
}

func (r *DeleteNoteRequest) Validate() error {
	if r == nil {
		return global.ErrNilReq
	}

	if r.NoteId <= 0 {
		return global.ErrArgs.Msg("笔记不存在")
	}

	return nil
}
