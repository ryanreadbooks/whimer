package systemnotify

import (
	"context"
	"encoding/json"

	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/event"
	"github.com/segmentio/kafka-go"
)

const (
	deletionTopic = "pilot_sysmsg_deletion_topic"
)

type EventPublisherImpl struct {
	asyncWriter *xkafka.Writer
	syncWriter  *xkafka.Writer
}

var _ event.EventPublisher = (*EventPublisherImpl)(nil)

func NewEventPublisherImpl(asyncWriter, syncWriter *xkafka.Writer) *EventPublisherImpl {
	return &EventPublisherImpl{
		asyncWriter: asyncWriter,
		syncWriter:  syncWriter,
	}
}

func (p *EventPublisherImpl) AsyncPublishDeletion(ctx context.Context, events []*event.DeletionEvent) error {
	if len(events) == 0 {
		return nil
	}

	msgs := makeDeletionEventKafkaMessages(events)
	ctx = context.WithoutCancel(ctx)
	return p.asyncWriter.WriteMessages(ctx, msgs...)
}

func (p *EventPublisherImpl) PublishDeletion(ctx context.Context, events []*event.DeletionEvent) error {
	if len(events) == 0 {
		return nil
	}

	msgs := makeDeletionEventKafkaMessages(events)
	return p.syncWriter.WriteMessages(ctx, msgs...)
}

type deletionEventPayload struct {
	MsgId string `json:"msg_id"`
	Uid   int64  `json:"uid"`
}

func makeDeletionEventKafkaMessages(evs []*event.DeletionEvent) []kafka.Message {
	msgs := make([]kafka.Message, 0, len(evs))

	for _, ev := range evs {
		payload, err := json.Marshal(deletionEventPayload{
			MsgId: ev.MsgId,
			Uid:   ev.Uid,
		})
		if err != nil {
			continue
		}

		msgs = append(msgs, kafka.Message{
			Topic: deletionTopic,
			Key:   []byte(ev.MsgId),
			Value: payload,
		})
	}

	return msgs
}
