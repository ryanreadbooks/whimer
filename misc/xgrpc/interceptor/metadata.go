package interceptor

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/metadata"

	"google.golang.org/grpc"
	grpcmeta "google.golang.org/grpc/metadata"
)

// 提取metadata中的请求上下文信息: 比如uid，设备信息等
func ServerMetadataExtract(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {

	md, ok := grpcmeta.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req)
	}

	for _, extractor := range metadata.RpcMdHolders {
		ctx = extractor.Extract(ctx, md)
	}

	return handler(ctx, req)
}

func ClientMetadataInject(ctx context.Context,
	method string,
	req, reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {

	// 从ctx中取出metadata并注入
	for _, injector := range metadata.RpcMdHolders {
		ctx = injector.Inject(ctx)
	}

	return invoker(ctx, method, req, reply, cc, opts...)
}
