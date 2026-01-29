package event

import "context"

type DeletionEvent struct {
	MsgId string
	Uid   int64
}

type EventPublisher interface {
	AsyncPublishDeletion(ctx context.Context, events []*DeletionEvent) error
	PublishDeletion(ctx context.Context, events []*DeletionEvent) error
}

