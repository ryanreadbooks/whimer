package metadata

import "context"

const (
	CtxUidKey        = "CtxUidKey"
	CtxSessIdKey     = "CtxSessIdKey"
	CtxClientAddrKey = "CtxClientAddrKey"
	CtxClientIpKey   = "CtxClientIpKey"
)

func getString(ctx context.Context, key string) string {
	v, _ := ctx.Value(key).(string)
	return v
}

func getUInt64(ctx context.Context, key string) uint64 {
	v, _ := ctx.Value(key).(uint64)
	return v
}

func WithUid(ctx context.Context, uid uint64) context.Context {
	return context.WithValue(ctx, CtxUidKey, uid)
}

func WithSessId(ctx context.Context, sessId string) context.Context {
	return context.WithValue(ctx, CtxSessIdKey, sessId)
}

func GetUid(ctx context.Context) uint64 {
	return getUInt64(ctx, CtxUidKey)
}

func GetSessId(ctx context.Context) string {
	return getString(ctx, CtxSessIdKey)
}

func WithClientAddr(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, CtxClientAddrKey, addr)
}

func GetClientAddr(ctx context.Context) string {
	return getString(ctx, CtxClientAddrKey)
}

func WithClientIp(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, CtxClientIpKey, ip)
}

func GetClientIp(ctx context.Context) string {
	return getString(ctx, CtxClientIpKey)
}
