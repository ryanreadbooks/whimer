package model

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/metadata"
)

const (
	CtxUserInfoKey = "__CtxUserInfoKey__"
	CtxSessIdKey   = "__CtxSessIdKey__"
)

func WithUserInfo(ctx context.Context, user *UserInfo) context.Context {
	ctx = context.WithValue(ctx, CtxUserInfoKey, user)
	return metadata.WithUid(ctx, user.Uid)
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
