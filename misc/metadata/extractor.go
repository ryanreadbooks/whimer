package metadata

import (
	"context"
	"strconv"

	grpcmetadata "google.golang.org/grpc/metadata"
)

type MdExtractor interface {
	Extract(ctx context.Context, md grpcmetadata.MD) context.Context
}

func (ij UidMdInOut) Extract(ctx context.Context, md grpcmetadata.MD) context.Context {
	res := md.Get(CtxUidKey)
	var uid int64
	if len(res) > 0 {
		uid, _ = strconv.ParseInt(res[0], 10, 64)
	}

	return WithUid(ctx, uid)
}

func (ij ClientIpInOut) Extract(ctx context.Context, md grpcmetadata.MD) context.Context {
	res := md.Get(CtxClientIpKey)
	var ip string
	if len(res) > 0 {
		ip = res[0]
	}

	return WithClientIp(ctx, ip)
}

func (ij ClientAddrInOut) Extract(ctx context.Context, md grpcmetadata.MD) context.Context {
	res := md.Get(CtxClientAddrKey)
	var addr string
	if len(res) > 0 {
		addr = res[0]
	}

	return WithClientAddr(ctx, addr)
}
