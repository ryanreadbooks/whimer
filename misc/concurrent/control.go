package concurrent

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"

	"github.com/panjf2000/ants/v2"
)

func ControllableExec[T any](ctx context.Context,
	pool *ants.Pool,
	datas []T,
	job func(ctx context.Context, datas []T) error) {

	SafeGo2(ctx, SafeGo2Opt{
		Name: "controllable_exec",
		Job: func(newCtx context.Context) error {
			// 借助pool来控制写入速度
			errSubmit := pool.Submit(func() {
				errExec := xslice.BatchExec(datas, 200, func(start, end int) error {
					errJob := job(newCtx, datas[start:end])
					if errJob != nil {
						return errJob
					}

					return nil
				})

				if errExec != nil {
					xlog.Msg("batch exec failed").Err(errExec).Errorx(newCtx)
				}
			})

			if errSubmit != nil {
				xlog.Msg("pool submit failed").Err(errSubmit).Errorx(newCtx)
				return errSubmit
			}

			return nil
		},
	})
}
