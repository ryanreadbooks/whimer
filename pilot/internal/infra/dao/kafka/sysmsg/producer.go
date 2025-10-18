package sysmsg

import (
	"context"
	"encoding/json"

	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"
	"github.com/segmentio/kafka-go"
)

type SysMsgProducer struct {
	asyncWriter *xkafka.Writer
	syncWriter  *xkafka.Writer
}

const (
	DeletionTopic      = "pilot_sysmsg_deletion_topic"
	DeletionTopicGroup = "pilot_sysmsg_deletion_consume_group"
)

// 系统消息lazy delete消息生产
func NewProducer(w, sw *xkafka.Writer) *SysMsgProducer {
	return &SysMsgProducer{
		asyncWriter: w,
		syncWriter:  sw,
	}
}

func (p *SysMsgProducer) AsyncPutDeletion(ctx context.Context, evs []*DeletionEvent) error {
	msgs := makeDeletionEventKafkaMessages(evs)

	ctx = context.WithoutCancel(ctx)
	return p.asyncWriter.WriteMessages(ctx, msgs...)
}

func (p *SysMsgProducer) PutDeletion(ctx context.Context, evs []*DeletionEvent) error {
	msgs := makeDeletionEventKafkaMessages(evs)

	return p.syncWriter.WriteMessages(ctx, msgs...)
}

func makeDeletionEventKafkaMessages(evs []*DeletionEvent) []kafka.Message {
	msgs := make([]kafka.Message, 0, len(evs))

	for _, ev := range evs {
		payload, err := json.Marshal(ev)
		if err != nil {
			continue
		}

		msgs = append(msgs, kafka.Message{
			Topic: DeletionTopic,
			Key:   []byte(ev.MsgId),
			Value: payload,
		})
	}

	return msgs
}
