package xcache

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func HGetAllWithScan(ctx context.Context, r *redis.Redis, key string, out any) error {
	pipe, err := r.TxPipeline()
	if err != nil {
		return err
	}

	cmd := pipe.HGetAll(ctx, key)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return err
	}

	result, err := cmd.Result()
	if err != nil {
		return err
	}

	if len(result) == 0 {
		return redis.Nil
	}

	return cmd.Scan(out)
}
