package dto

import (
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/noteinteract/errors"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"
)

type LikeNoteCommand struct {
	NoteId vo.NoteId     `json:"note_id"`
	Action vo.LikeAction `json:"action"`
}

func (c *LikeNoteCommand) Validate() error {
	if c == nil {
		return xerror.ErrNilArg
	}

	if c.Action != vo.LikeActionDo && c.Action != vo.LikeActionUndo {
		return errors.ErrUnsupportedAction
	}

	return nil
}
