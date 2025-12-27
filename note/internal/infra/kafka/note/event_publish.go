package note

import (
	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"
)

// 笔记事件发布

type EventPublisher struct {
	w *xkafka.Writer
}

func NewEventPublisher(w *xkafka.Writer) *EventPublisher {
	return &EventPublisher{
		w: w,
	}
}
