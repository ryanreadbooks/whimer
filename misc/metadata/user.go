package metadata

import "context"

const (
	CtxUidKey        = "CtxUidKey"
	CtxSessIdKey     = "CtxSessIdKey"
	CtxClientAddrKey = "CtxClientAddrKey"
	CtxClientIpKey   = "CtxClientIpKey"
)

func WithUid(ctx context.Context, uid uint64) context.Context {
	return context.WithValue(ctx, CtxUidKey, uid)
}

func WithSessId(ctx context.Context, sessId string) context.Context {
	return context.WithValue(ctx, CtxSessIdKey, sessId)
}

func Uid(ctx context.Context) uint64 {
	return getUInt64(ctx, CtxUidKey)
}

func SessId(ctx context.Context) string {
	return getString(ctx, CtxSessIdKey)
}

func WithClientAddr(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, CtxClientAddrKey, addr)
}

func ClientAddr(ctx context.Context) string {
	return getString(ctx, CtxClientAddrKey)
}

func WithClientIp(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, CtxClientIpKey, ip)
}

func ClientIp(ctx context.Context) string {
	return getString(ctx, CtxClientIpKey)
}
