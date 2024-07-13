package model

import (
	"context"

	"github.com/ryanreadbooks/whimer/passport/internal/model/profile"
)

const (
	CtxMeInfoKey = "CtxMeInfoKey"
	CtxSessIdKey = "CtxSessIdKey"
)

func WithMeInfo(ctx context.Context, me *profile.MeInfo) context.Context {
	return context.WithValue(ctx, CtxMeInfoKey, me)
}

func WithSessId(ctx context.Context, sessId string) context.Context {
	return context.WithValue(ctx, CtxSessIdKey, sessId)
}

func CtxGetMeInfo(ctx context.Context) *profile.MeInfo {
	if me, ok := ctx.Value(CtxMeInfoKey).(*profile.MeInfo); ok {
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
