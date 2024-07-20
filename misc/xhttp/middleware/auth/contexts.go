package auth

import "context"

const (
	CtxUidKey    = "CtxUidKey"
	CtxSessIdKey = "CtxSessIdKey"
)

func WithUid(ctx context.Context, uid uint64) context.Context {
	return context.WithValue(ctx, CtxUidKey, uid)
}

func WithSessId(ctx context.Context, sessId string) context.Context {
	return context.WithValue(ctx, CtxSessIdKey, sessId)
}

func CtxGetUid(ctx context.Context) uint64 {
	v, _ := ctx.Value(CtxUidKey).(uint64)
	return v
}

func CtxGetSessId(ctx context.Context) string {
	v, _ := ctx.Value(CtxSessIdKey).(string)
	return v
}
