package concurrent

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

func SafeGo(job func()) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logx.ErrorStackf("panic: %v", err)
			}
		}()

		job()
	}()
}

type DoneInJob func(ctx context.Context)

func DoneIn(duration time.Duration, job DoneInJob) {
	DoneInCtx(context.Background(), duration, job)
}

func DoneInCtx(parent context.Context, duration time.Duration, job DoneInJob) {
	SafeGo(func() {
		parent = context.WithoutCancel(parent)
		ctx, cancel := context.WithTimeout(parent, duration)
		defer cancel()

		job(ctx)
	})
}
