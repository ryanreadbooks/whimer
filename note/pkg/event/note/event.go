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

type Note = event.Note

type (
	NoteEvent              = event.NoteEvent
	NotePublishedEventData = event.NotePublishedEventData
	NoteDeletedEventData   = event.NoteDeletedEventData
	NoteRejectedEventData  = event.NoteRejectedEventData
	NoteBannedEventData    = event.NoteBannedEventData
	NoteLikedEventData     = event.NoteLikedEventData
	NoteCommentedEventData = event.NoteCommentedEventData
)
