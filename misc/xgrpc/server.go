package xgrpc

import (
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func EnableReflectionIfNecessary(c zrpc.RpcServerConf, s *grpc.Server) {
	if c.Mode == service.DevMode || c.Mode == service.TestMode {
		reflection.Register(s)
	}
}
