package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	infrakfk "github.com/ryanreadbooks/whimer/note/internal/infra/kafka"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	eventmodel "github.com/ryanreadbooks/whimer/note/internal/model/event"
	pkgid "github.com/ryanreadbooks/whimer/note/pkg/id"

	"github.com/segmentio/kafka-go"
)

const (
	NoteEventTopic = "note_event"
)

// 笔记相关事件统一处理
type NoteEventBus struct {
	pub *infrakfk.Publisher
}

func NewNoteEventBus(pub *infrakfk.Publisher) *NoteEventBus {
	return &NoteEventBus{
		pub: pub,
	}
}

func (e *NoteEventBus) makeNoteEvent(ctx context.Context,
	note *model.Note,
	eventType eventmodel.EventType,
	payload any,
) (string, []byte, error) {
	now := time.Now()
	noteId := pkgid.NoteId(note.NoteId).String()
	evt := &eventmodel.NoteEvent{
		Type:      eventType,
		NoteId:    noteId,
		Timestamp: now.UnixMilli(),
		Payload:   payload,
	}

	evtBytes, err := json.Marshal(evt)
	if err != nil {
		return "", nil, xerror.Wrapf(err, "note event bus make note event failed to marshal event").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	return noteId, evtBytes, nil
}

// 笔记发布
func (e *NoteEventBus) NotePublished(ctx context.Context, note *model.Note) error {
	noteId, evtBytes, err := e.makeNoteEvent(ctx,
		note,
		eventmodel.NotePublished,
		&eventmodel.NotePublishedEventData{
			Note: modelNoteToEventNote(note),
		})
	if err != nil {
		return xerror.Wrapf(err, "note event bus note published failed to make note event").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	err = e.pub.Writer().WriteMessages(ctx, kafka.Message{
		Topic: NoteEventTopic,
		Key:   []byte(noteId),
		Value: evtBytes,
	})
	if err != nil {
		return xerror.Wrapf(err, "note event bus note published failed to write message").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	return nil
}

func (e *NoteEventBus) NoteDeleted(ctx context.Context, note *model.Note, reason eventmodel.NoteDeleteReason) error {
	noteId, evtBytes, err := e.makeNoteEvent(ctx,
		note,
		eventmodel.NoteDeleted,
		&eventmodel.NoteDeletedEventData{
			Note:   modelNoteToEventNote(note),
			Reason: reason,
		})
	if err != nil {
		return xerror.Wrapf(err, "note event bus note deleted failed to make note event").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	err = e.pub.Writer().WriteMessages(ctx, kafka.Message{
		Topic: NoteEventTopic,
		Key:   []byte(noteId),
		Value: evtBytes,
	})
	if err != nil {
		return xerror.Wrapf(err, "note event bus note deleted failed to write message").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	return nil
}
