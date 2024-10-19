package interceptor

import (
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

// server端拦截器
func InstallServerUnaryInterceptors(server *zrpc.RpcServer,
	customs ...grpc.UnaryServerInterceptor) {
	// 默认拦截器
	interceptors := []grpc.UnaryServerInterceptor{
		ServerErrorHandle,
		ServerMetadataExtract,
		ServerValidateHandle,
	}

	// 自定义拦截器
	interceptors = append(interceptors, customs...)
	server.AddUnaryInterceptors(interceptors...)
}

func WithChecker(checkers ...ServerMetadateChecker) grpc.UnaryServerInterceptor {
	return ServerMetadataCheck(checkers...)
}
