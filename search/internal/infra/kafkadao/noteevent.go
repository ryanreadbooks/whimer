package kafkadao

import (
	"context"
	"encoding/json"

	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	searchv1 "github.com/ryanreadbooks/whimer/search/api/v1"
	"github.com/segmentio/kafka-go"

	"google.golang.org/protobuf/encoding/protojson"
)

type NoteEventType string

const (
	NoteAddEvent     NoteEventType = "note_add"     // 添加笔记事件
	NoteUpdateEvent  NoteEventType = "note_update"  // 更新笔记事件
	NoteDeleteEvent  NoteEventType = "note_delete"  // 删除笔记事件
	NoteLikeEvent    NoteEventType = "note_like"    // 笔记赞事件
	NoteCommentEvent NoteEventType = "note_comment" // 笔记评论事件
)

type NoteEvent struct {
	Type    NoteEventType   `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

const (
	EsNoteTopic      = "es_note_events"       // 笔记相关事件主题
	EsNoteTopicGroup = "es_note_events_group" // 笔记相关事件消费者组名称
)

type NoteLikeEventPayload struct {
	NoteId    string `json:"note_id"`
	Increment int64  `json:"increment"`
}

type NoteCommentEventPayload struct {
	NoteId    string `json:"note_id"`
	Increment int64  `json:"increment"`
}

type NoteEventProducer struct {
	w *xkafka.Writer
}

func (p *NoteEventProducer) PutNoteAddEvent(ctx context.Context, evs []*searchv1.Note) error {
	msgs := make([]kafka.Message, 0, len(evs))

	for _, ev := range evs {
		payload, err := protojson.Marshal(ev)
		if err != nil {
			xlog.Msg("failed to protojson marshal req").Err(err).Errorx(ctx)
			continue
		}

		event := &NoteEvent{
			Type:    NoteAddEvent,
			Payload: payload,
		}

		value, err := json.Marshal(event)
		if err != nil {
			xlog.Msg("failed to json marshal add event").Err(err).Errorx(ctx)
			continue
		}

		msgs = append(msgs, kafka.Message{
			Topic: EsNoteTopic,
			Key:   []byte(ev.NoteId),
			Value: value,
		})
	}

	return p.w.WriteMessages(ctx, msgs...) // async will always return nil error
}

func (p *NoteEventProducer) PutNoteDeleteEvent(ctx context.Context, evs []string) error {
	msgs := make([]kafka.Message, 0, len(evs))

	for _, noteId := range evs {
		payload, err := json.Marshal(noteId)
		if err != nil {
			xlog.Msg("failed to json marshal req").Err(err).Errorx(ctx)
			continue
		}

		event := &NoteEvent{
			Type:    NoteDeleteEvent,
			Payload: payload,
		}
		value, err := json.Marshal(event)
		if err != nil {
			xlog.Msg("failed to json marshal delete event").Err(err).Errorx(ctx)
			continue
		}

		msgs = append(msgs, kafka.Message{
			Topic: EsNoteTopic,
			Key:   []byte(noteId), // noteId
			Value: value,
		})
	}

	return p.w.WriteMessages(ctx, msgs...)
}

// reqs: note_id -> like_count increment
func (p *NoteEventProducer) PutNoteLikeEvent(ctx context.Context, reqs map[string]int64) error {
	msgs := make([]kafka.Message, 0)
	for noteId, increment := range reqs {
		payload, err := json.Marshal(&NoteLikeEventPayload{
			NoteId:    noteId,
			Increment: increment,
		})
		if err != nil {
			xlog.Msg("failed to json marshal like count payload").Err(err).Errorx(ctx)
			continue
		}

		event := &NoteEvent{
			Type:    NoteLikeEvent,
			Payload: payload,
		}
		value, err := json.Marshal(event)
		if err != nil {
			xlog.Msg("failed to json marshal like event").Err(err).Errorx(ctx)
			continue
		}

		msgs = append(msgs, kafka.Message{
			Topic: EsNoteTopic,
			Key:   []byte(noteId),
			Value: value,
		})
	}

	return p.w.WriteMessages(ctx, msgs...)
}

// reqs: note_id -> comment_count increment
func (p *NoteEventProducer) PutNoteCommentEvent(ctx context.Context, reqs map[string]int64) error {
	msgs := make([]kafka.Message, 0)
	for noteId, increment := range reqs {
		payload, err := json.Marshal(&NoteCommentEventPayload{
			NoteId:    noteId,
			Increment: increment,
		})
		if err != nil {
			xlog.Msg("failed to json marshal comment count payload").Err(err).Errorx(ctx)
			continue
		}

		event := &NoteEvent{
			Type:    NoteCommentEvent,
			Payload: payload,
		}
		value, err := json.Marshal(event)
		if err != nil {
			xlog.Msg("failed to json marshal comment event").Err(err).Errorx(ctx)
			continue
		}

		msgs = append(msgs, kafka.Message{
			Topic: EsNoteTopic,
			Key:   []byte(noteId),
			Value: value,
		})
	}

	return p.w.WriteMessages(ctx, msgs...)
}
