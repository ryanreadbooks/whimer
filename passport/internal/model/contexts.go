package model

import "context"

const (
	CtxMeInfoKey = "CtxMeInfoKey"
)

func WithMeInfo(ctx context.Context, me *MeInfo) context.Context {
	return context.WithValue(ctx, CtxMeInfoKey, me)
}

func CtxGetMeInfo(ctx context.Context) *MeInfo {
	if me, ok := ctx.Value(CtxMeInfoKey).(*MeInfo); ok {
		return me
	}

	return nil
}
