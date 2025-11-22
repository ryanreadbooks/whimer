package interceptor

import (
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor/checker"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

// server端拦截器
func InstallUnaryServerInterceptors(server *zrpc.RpcServer,
	customs ...grpc.UnaryServerInterceptor) {
	// 默认拦截器
	interceptors := []grpc.UnaryServerInterceptor{
		UnaryServerErrorHandler,
		UnaryServerRecovery,
		UnaryServerMetadataExtract,
		UnaryServerValidateHandler,
		UnaryServerExtensionHandler,
	}

	// 自定义拦截器
	interceptors = append(interceptors, customs...)
	server.AddUnaryInterceptors(interceptors...)
}

func WithUnaryChecker(checkers ...checker.UnaryServerMetadataChecker) grpc.UnaryServerInterceptor {
	return UnaryServerMetadataCheck(checkers...)
}
