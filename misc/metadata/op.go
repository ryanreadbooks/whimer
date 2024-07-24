package metadata

import (
	"context"
)

func getString(ctx context.Context, key string) string {
	v, _ := ctx.Value(key).(string)
	return v
}

func getUInt64(ctx context.Context, key string) uint64 {
	v, _ := ctx.Value(key).(uint64)
	return v
}
