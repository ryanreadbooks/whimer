package checker

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/metadata"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/util"

	"google.golang.org/grpc"
)

type UnaryServerMetadataChecker func(ctx context.Context, info *grpc.UnaryServerInfo) error

func UidExistence(ctx context.Context, info *grpc.UnaryServerInfo) error {
	uid := metadata.Uid(ctx)
	if uid <= 0 {
		return xerror.ErrNotLogin
	}

	return nil
}

type option struct {
	ignores []string
}

type Option func(o *option)

func WithIgnore(ignores ...string) Option {
	return func(o *option) {
		o.ignores = ignores
	}
}

// 可设置某些服务忽略此中间件检查
func UidExistenceWithOpt(opts ...Option) UnaryServerMetadataChecker {
	var opt option
	for _, o := range opts {
		o(&opt)
	}

	var ignoreMap = make(map[string]struct{}, len(opt.ignores))
	for _, e := range opt.ignores {
		ignoreMap[e] = struct{}{}
	}

	// true: should ignore; false should not ignore
	shouldIgnore := func(info *grpc.UnaryServerInfo) bool {
		svcName := util.SplitUnaryServerName(info)
		_, ok := ignoreMap[svcName]
		return ok
	}

	// interceptor
	return func(ctx context.Context, info *grpc.UnaryServerInfo) error {
		if shouldIgnore(info) {
			return nil
		}

		uid := metadata.Uid(ctx)
		if uid <= 0 {
			return xerror.ErrNotLogin
		}

		return nil
	}
}
