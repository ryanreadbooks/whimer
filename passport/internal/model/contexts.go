package model

import "context"

const (
	CtxMeInfoKey = "CtxMeInfoKey"
	CtxSessIdKey = "CtxSessIdKey"
)

func WithMeInfo(ctx context.Context, me *MeInfo) context.Context {
	return context.WithValue(ctx, CtxMeInfoKey, me)
}

func WithSessId(ctx context.Context, sessId string) context.Context {
	return context.WithValue(ctx, CtxSessIdKey, sessId)
}

func CtxGetMeInfo(ctx context.Context) *MeInfo {
	if me, ok := ctx.Value(CtxMeInfoKey).(*MeInfo); ok {
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
