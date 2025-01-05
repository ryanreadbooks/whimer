package xkq

import (
	"context"
	"fmt"

	"github.com/ryanreadbooks/whimer/misc/stacktrace"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

type Consumer func(ctx context.Context, key, value string) error

func (f Consumer) Consume(ctx context.Context, key, value string) (err error) {
	defer func() {
		var log *xlog.LogItem
		if err != nil {
			log = xlog.Msg("consumer error").Err(err)
		}
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
			log = xlog.Msg("panic during consume").
				Err(err).
				Extra("trace", stacktrace.FormatFrames(stacktrace.TraceStack()))
		}
		if err != nil && log != nil {
			log.Errorx(ctx)
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
