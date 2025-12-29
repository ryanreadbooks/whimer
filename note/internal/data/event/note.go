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

// 笔记发布
func (e *NoteEventBus) NotePublished(ctx context.Context, note *model.Note) error {
	now := time.Now()
	// TODO 按照note生成payload
	payload := &eventmodel.NotePublishedEventData{}

	noteId := pkgid.NoteId(note.NoteId).String()
	evt := &eventmodel.NoteEvent{
		Type:      eventmodel.NotePublished,
		NoteId:    noteId,
		Timestamp: now.UnixMilli(),
		Payload:   payload,
	}

	evtBytes, err := json.Marshal(evt)
	if err != nil {
		return xerror.Wrapf(err, "note event bus note published failed to marshal event").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	err = e.pub.Writer().WriteMessages(ctx, kafka.Message{
		Topic: NoteEventTopic,
		Key:   []byte(noteId),
		Value: evtBytes,
		Time:  now,
	})
	if err != nil {
		return xerror.Wrapf(err, "note event bus note published failed to write message").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	return nil
}
