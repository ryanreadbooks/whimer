package xkq

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/stacktrace"
)


type Consumer func(ctx context.Context, key, value string) error

func (f Consumer) Consume(ctx context.Context, key, value string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("panic during Consume: %v, trace: %v", e, stacktrace.FormatFrames(stacktrace.TraceStack()))
		}
	}()
	
	return f(ctx, key, value)
}

type KeyMatcher interface {
	Match(key string) Consumer
}

type ValueDeserializer[T any] interface {
	Deserialize(value string) T
}