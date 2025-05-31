package metadata

import (
	"context"
	"strconv"

	grpcmetadata "google.golang.org/grpc/metadata"
)

type GrpcMdHolder interface {
	MdInjector
	MdExtractor
}

type MdInjector interface {
	// 从ctx中取出value放入metadata中并重新设置回ctx中
	Inject(ctx context.Context) context.Context
}

type UidMdInOut struct{}

func (ij UidMdInOut) Inject(ctx context.Context) context.Context {
	uid := Uid(ctx)
	return grpcmetadata.AppendToOutgoingContext(ctx, CtxUidKey, strconv.FormatInt(uid, 10))
}

type ClientIpInOut struct{}

func (ij ClientIpInOut) Inject(ctx context.Context) context.Context {
	ip := ClientIp(ctx)
	if len(ip) > 0 {
		return grpcmetadata.AppendToOutgoingContext(ctx, CtxClientIpKey, ip)
	}
	return ctx
}

type ClientAddrInOut struct{}

func (ij ClientAddrInOut) Inject(ctx context.Context) context.Context {
	addr := ClientAddr(ctx)
	if len(addr) > 0 {
		return grpcmetadata.AppendToOutgoingContext(ctx, CtxClientAddrKey, addr)
	}
	return ctx
}

var (
	RpcMdHolders = []GrpcMdHolder{
		UidMdInOut{},
		ClientIpInOut{},
		ClientAddrInOut{},
	}
)
