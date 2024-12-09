package checker

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xlog"

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

func UidExistenceLoose(ctx context.Context, info *grpc.UnaryServerInfo) error {
	uid := metadata.Uid(ctx)
	if uid <= 0 {
		xlog.Msg("uid not found in grpc incoming metadata").Info()
	}

	return nil
}

type option struct {
	ignoredServices []string // 不生效的服务
	ignoredMethods  []string // 不生效的方法
}

type Option func(o *option)

// 拦截器对某些grpc服务不生效
func WithServicesIgnore(ignores ...string) Option {
	return func(o *option) {
		o.ignoredServices = ignores
	}
}

// 拦截器对某给grpc方法不生效
func WithMethodsIgnore(methods ...string) Option {
	return func(o *option) {
		o.ignoredMethods = methods
	}
}

// 可设置某些服务忽略此中间件检查
func UidExistenceWithOpt(opts ...Option) UnaryServerMetadataChecker {
	var opt option
	for _, o := range opts {
		o(&opt)
	}

	var ignoreServiceMap = make(map[string]struct{}, len(opt.ignoredServices))
	for _, e := range opt.ignoredServices {
		ignoreServiceMap[e] = struct{}{}
	}
	var ignoreMethodMap = make(map[string]struct{}, len(opt.ignoredMethods))
	for _, e := range opt.ignoredMethods {
		ignoreMethodMap[e] = struct{}{}
	}

	// true: should ignore; false should not ignore
	shouldIgnore := func(info *grpc.UnaryServerInfo) bool {
		// rules: 1. service first; 2. if service is not ignored then check method
		svcName := util.SplitUnaryServerName(info)
		_, ok := ignoreServiceMap[svcName]
		if ok {
			return ok
		}

		_, ok = ignoreMethodMap[info.FullMethod]
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
