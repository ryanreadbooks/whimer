package note

import (
	"encoding/json"
	"fmt"
)

type NoteActionType int8

const (
	ActionCreateNote             NoteActionType = 1
	ActionDeleteNote             NoteActionType = 2
	ActionUpdateNoteLikeCount    NoteActionType = 3
	ActionUpdateNoteCommentCount NoteActionType = 4
)

type NoteAction interface {
	Type() NoteActionType
	GetDocId() string
	GetDoc() (any, error)
}

type noteCreateAction struct {
	data *Note
}

func (n *noteCreateAction) Type() NoteActionType { return ActionCreateNote }

func (n *noteCreateAction) GetDoc() (any, error) { return json.Marshal(n.data) }

func (n *noteCreateAction) GetDocId() string { return n.data.GetId() }

func NewNoteCreateAction(n *Note) *noteCreateAction {
	ac := noteCreateAction{
		data: n,
	}

	return &ac
}

type noteDeleteAction struct {
	noteId string
}

func (n *noteDeleteAction) Type() NoteActionType { return ActionDeleteNote }

func (n *noteDeleteAction) GetDoc() (any, error) { return nil, fmt.Errorf("nil doc") }

func (n *noteDeleteAction) GetDocId() string {
	return fmtNoteDocIdString(n.noteId)
}

func NewNoteDeleteAction(noteId string) *noteDeleteAction {
	return &noteDeleteAction{noteId: noteId}
}

type noteUpdateLikeCountAction struct {
	noteId string
	incr   int64
}

func (n *noteUpdateLikeCountAction) Type() NoteActionType { return ActionUpdateNoteLikeCount }

func (n *noteUpdateLikeCountAction) GetDoc() (any, error) {
	return n.incr, nil
}

func (n *noteUpdateLikeCountAction) GetDocId() string {
	return fmtNoteDocIdString(n.noteId)
}

func NewNoteUpdateLikeCountAction(noteId string, incr int64) *noteUpdateLikeCountAction {
	return &noteUpdateLikeCountAction{
		noteId: noteId,
		incr:   incr,
	}
}

type noteUpdateCommentCountAction struct {
	noteId string
	incr   int64
}

func (n *noteUpdateCommentCountAction) Type() NoteActionType { return ActionUpdateNoteLikeCount }

func (n *noteUpdateCommentCountAction) GetDoc() (any, error) {
	return n.incr, nil
}

func (n *noteUpdateCommentCountAction) GetDocId() string {
	return fmtNoteDocIdString(n.noteId)
}

func NewNoteUpdateCommentCountAction(noteId string, incr int64) *noteUpdateCommentCountAction {
	return &noteUpdateCommentCountAction{
		noteId: noteId,
		incr:   incr,
	}
}
