package note

import (
	"github.com/ryanreadbooks/whimer/note/internal/model/event"
)

type EventType = event.EventType

const (
	NotePublished = event.NotePublished
	NoteRejected  = event.NoteRejected
	NoteBanned    = event.NoteBanned
	NoteLiked     = event.NoteLiked
	NoteCommented = event.NoteCommented
)
