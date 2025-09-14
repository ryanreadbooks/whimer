package functions

import (
	"context"
	_ "embed"

	goredis "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

//go:embed lua/global_fn.lua
var globalFnLua string

func GetLibFunctions() string {
	return globalFnLua
}

func FunctionLoadReplace(ctx context.Context, r *redis.Redis, code string) error {
	return r.PipelinedCtx(ctx, func(p redis.Pipeliner) error {
		p.FunctionLoadReplace(ctx, code)
		return nil
	})
}

func FunctionCall(ctx context.Context, r *redis.Redis, fn string, keys []string, args ...any) (*goredis.Cmd, error) {
	pipe, err := r.TxPipeline()
	if err != nil {
		return nil, err
	}

	cmd := pipe.FCall(ctx, fn, keys, args...)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	return cmd, nil
}
