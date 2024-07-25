package sdk

import (
	"context"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type Note interface {
	IsUserOwnNote(ctx context.Context, in *IsUserOwnNoteReq, opts ...grpc.CallOption) (*IsUserOwnNoteRes, error)
}

type defaultNote struct {
	cli zrpc.Client
}

func NewNote(cli zrpc.Client) Note {
	return &defaultNote{
		cli: cli,
	}
}

func (m *defaultNote) IsUserOwnNote(ctx context.Context, in *IsUserOwnNoteReq, opts ...grpc.CallOption) (*IsUserOwnNoteRes, error) {
	client := NewNoteClient(m.cli.Conn())
	return client.IsUserOwnNote(ctx, in, opts...)
}
