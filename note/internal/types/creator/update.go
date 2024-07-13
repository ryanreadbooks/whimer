package creator

import (
	"github.com/ryanreadbooks/whimer/note/internal/global"
)

type UpdateReq struct {
	NoteId string `json:"note_id"`
	CreateReq
}

func (r *UpdateReq) Validate() error {
	if r == nil {
		return global.ErrNilReq
	}

	if len(r.NoteId) == 0 {
		return global.ErrArgs.Msg("笔记id错误")
	}

	return r.CreateReq.Validate()
}

type UpdateRes struct {
	NoteId string `json:"note_id"`
}
