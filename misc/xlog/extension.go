package xlog

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/zeromicro/go-zero/core/logx"
)

func WithUid(ctx context.Context) logx.LogField {
	return logx.LogField{Key: "uid", Value: metadata.Uid(ctx)}
}

func WithErr(err error) logx.LogField {
	return logx.LogField{Key: "err", Value: err.Error()}
}
