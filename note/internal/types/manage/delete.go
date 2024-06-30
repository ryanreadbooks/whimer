package manage

import "github.com/ryanreadbooks/whimer/note/internal/global"

type DeleteReq struct {
	NoteId string `json:"note_id"`
}

func (r *DeleteReq) Validate() error {
	if r == nil {
		return global.ErrNilReq
	}

	if len(r.NoteId) <= 0 {
		return global.ErrArgs.Msg("笔记不存在")
	}

	return nil
}
