package interceptor

import (
	"context"
	"strconv"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"google.golang.org/grpc"
	grpcmeta "google.golang.org/grpc/metadata"
)

// 提取metadata中的请求上下文信息: 比如uid，设备信息等
func ServerMetadataHandle(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {

	md, ok := grpcmeta.FromIncomingContext(ctx)
	if !ok {
		return handler(ctx, req)
	}

	res := md.Get(metadata.CtxUidKey)
	var uid uint64
	if len(res) > 0 {
		uid, _ = strconv.ParseUint(res[0], 10, 64)
	}
	ctx = metadata.WithUid(ctx, uid)

	res = md.Get(metadata.CtxClientIpKey)
	var ip string
	if len(res) > 0 {
		ip = res[0]
	}
	ctx = metadata.WithClientIp(ctx, ip)

	res = md.Get(metadata.CtxClientAddrKey)
	var addr string
	if len(res) > 0 {
		addr = res[0]
	}
	ctx = metadata.WithClientAddr(ctx, addr)

	return handler(ctx, req)
}
