package model

import (
	"context"
)

const (
	CtxUserInfoKey = "__CtxUserInfoKey__"
	CtxSessIdKey   = "__CtxSessIdKey__"
)

func WithUserInfo(ctx context.Context, me *UserInfo) context.Context {
	return context.WithValue(ctx, CtxUserInfoKey, me)
}

func WithSessId(ctx context.Context, sessId string) context.Context {
	return context.WithValue(ctx, CtxSessIdKey, sessId)
}

func CtxGetUserInfo(ctx context.Context) *UserInfo {
	if me, ok := ctx.Value(CtxUserInfoKey).(*UserInfo); ok {
		return me
	}

	return nil
}

func CtxGetSessId(ctx context.Context) string {
	if sessId, ok := ctx.Value(CtxSessIdKey).(string); ok {
		return sessId
	}
	return ""
}
