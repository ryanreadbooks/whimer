package vo

import (
	noteid "github.com/ryanreadbooks/whimer/note/pkg/id"
)

type NotifyLikesOnNoteParam struct {
	NoteId noteid.NoteId `json:"note_id"`
}

type NotifyLikesOnCommentParam struct {
	NoteId    noteid.NoteId `json:"note_id"`
	CommentId int64         `json:"comment_id"`
}
