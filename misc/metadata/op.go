package metadata

import (
	"context"
)

func getString(ctx context.Context, key string) string {
	v, _ := ctx.Value(key).(string)
	return v
}

func getInt64(ctx context.Context, key string) int64 {
	v, _ := ctx.Value(key).(int64)
	return v
}
