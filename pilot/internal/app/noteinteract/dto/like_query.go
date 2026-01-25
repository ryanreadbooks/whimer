package dto

import "github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"

type GetLikeCountResult struct {
	NoteId vo.NoteId `json:"note_id"`
	Count  int64     `json:"count"`
}
