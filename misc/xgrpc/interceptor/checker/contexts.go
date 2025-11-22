package checker

import "context"

type skipUidCheckCtxKey struct{}

func ForceSkipUidCheck(ctx context.Context) context.Context {
	return context.WithValue(ctx, skipUidCheckCtxKey{}, true)
}

func GetForceSkipUidCheck(ctx context.Context) bool {
	v, _ := ctx.Value(skipUidCheckCtxKey{}).(bool)
	return v
}
