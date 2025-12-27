package kafka

import (
	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"
	"github.com/ryanreadbooks/whimer/note/internal/infra/kafka/note"
)

type Publisher struct {
	NoteEvent *note.EventPublisher
}

func New(writer, asyncWriter *xkafka.Writer) *Publisher {
	return &Publisher{
		NoteEvent: note.NewEventPublisher(asyncWriter),
	}
}
