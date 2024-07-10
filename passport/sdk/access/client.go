package access

import (
	"context"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type Access interface {
	CheckSignIn(ctx context.Context, in *CheckSignInReq, opts ...grpc.CallOption) (*CheckSignInRes, error)
}

type defaultAccess struct {
	cli zrpc.Client
}

func NewAccess(cli zrpc.Client) Access {
	return &defaultAccess{
		cli: cli,
	}
}

func (m *defaultAccess) CheckSignIn(ctx context.Context, in *CheckSignInReq, opts ...grpc.CallOption) (*CheckSignInRes, error) {
	client := NewAccessClient(m.cli.Conn())
	return client.CheckSignIn(ctx, in, opts...)
}
