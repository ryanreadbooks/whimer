package concur

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

func DoneIn(duraton time.Duration, job func(ctx context.Context)) {
	SafeGo(func() {
		ctx, cancel := context.WithTimeout(context.Background(), duraton)
		defer cancel()

		job(ctx)
	})
}