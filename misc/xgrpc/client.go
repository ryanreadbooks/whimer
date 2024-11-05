package xgrpc

import (
	"github.com/ryanreadbooks/whimer/misc/xconf"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

// 快速创建带通用拦截器的grpc客户端连接
func NewClientFromDiscovery(conf xconf.Discovery) (zrpc.Client, error) {
	cli, err := zrpc.NewClient(
		conf.AsZrpcClientConf(),
		zrpc.WithUnaryClientInterceptor(interceptor.UnaryClientMetadataInject),
	)

	return cli, err
}

// 泛型版本
// 快速创建带通用来拦截器的grpc客户端连接
func NewClient[T any, P func(cc grpc.ClientConnInterface) T](
	conf xconf.Discovery, constructor P) (ret T, err error) {

	cli, err := zrpc.NewClient(
		conf.AsZrpcClientConf(),
		zrpc.WithUnaryClientInterceptor(interceptor.UnaryClientErrorHandler),
		zrpc.WithUnaryClientInterceptor(interceptor.UnaryClientMetadataInject),
	)

	if err != nil {
		return
	}

	return constructor(cli.Conn()), nil
}

// 创建带通用拦截器的grpc客户端连接
func NewClientConn(conf xconf.Discovery) (*grpc.ClientConn, error) {
	cli, err := zrpc.NewClient(
		conf.AsZrpcClientConf(),
		zrpc.WithUnaryClientInterceptor(interceptor.UnaryClientErrorHandler),
		zrpc.WithUnaryClientInterceptor(interceptor.UnaryClientMetadataInject),
	)
	if err != nil {
		return nil, err
	}

	return cli.Conn(), nil
}
