package metadata

import (
	"context"
	"strconv"

	grpcmeta "google.golang.org/grpc/metadata"
)

const (
	CtxUidKey          = "CtxUidKey"
	CtxUserNicknameKey = "CtxUserNicknameKey"
	CtxSessIdKey       = "CtxSessIdKey"
	CtxClientAddrKey   = "CtxClientAddrKey"
	CtxClientIpKey     = "CtxClientIpKey"
)

func WithUid(ctx context.Context, uid int64) context.Context {
	return context.WithValue(ctx, CtxUidKey, uid)
}

func HasUid(ctx context.Context) bool {
	return Uid(ctx) != 0
}

func WithUserNickname(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, CtxUserNicknameKey, name)
}

func HasUserNickname(ctx context.Context) bool {
	return UserNickname(ctx) != ""
}

func RpcWithUid(ctx context.Context, uid int64) context.Context {
	return grpcmeta.AppendToOutgoingContext(ctx, CtxUidKey, strconv.FormatInt(uid, 10))
}

func WithSessId(ctx context.Context, sessId string) context.Context {
	return context.WithValue(ctx, CtxSessIdKey, sessId)
}

func HasSessId(ctx context.Context) bool {
	return SessId(ctx) != ""
}

func RpcWithSessId(ctx context.Context, sessId string) context.Context {
	return grpcmeta.AppendToOutgoingContext(ctx, CtxSessIdKey, sessId)
}

func Uid(ctx context.Context) int64 {
	return getInt64(ctx, CtxUidKey)
}

func UserNickname(ctx context.Context) string {
	return getString(ctx, CtxUserNicknameKey)
}

func SessId(ctx context.Context) string {
	return getString(ctx, CtxSessIdKey)
}

func WithClientAddr(ctx context.Context, addr string) context.Context {
	return context.WithValue(ctx, CtxClientAddrKey, addr)
}

func RpcWithClientAddr(ctx context.Context, addr string) context.Context {
	return grpcmeta.AppendToOutgoingContext(ctx, CtxClientAddrKey, addr)
}

func ClientAddr(ctx context.Context) string {
	return getString(ctx, CtxClientAddrKey)
}

func WithClientIp(ctx context.Context, ip string) context.Context {
	return context.WithValue(ctx, CtxClientIpKey, ip)
}

func RpcWithClientIp(ctx context.Context, ip string) context.Context {
	return grpcmeta.AppendToOutgoingContext(ctx, CtxClientIpKey, ip)
}

func ClientIp(ctx context.Context) string {
	return getString(ctx, CtxClientIpKey)
}
