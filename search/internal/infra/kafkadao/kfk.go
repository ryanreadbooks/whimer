package kafkadao

import (
	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"
)

type KafkaDao struct {
	w                 *xkafka.Writer
	NoteEventProducer *NoteEventProducer
}

func New(w *xkafka.Writer) *KafkaDao {
	return &KafkaDao{
		w:                 w,
		NoteEventProducer: &NoteEventProducer{w: w},
	}
}
