package sdk

import (
	"context"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type Note interface {
	NoteClient
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

func (m *defaultNote) GetNote(ctx context.Context, in *GetNoteReq, opts ...grpc.CallOption) (*GetNoteRes, error) {
	client := NewNoteClient(m.cli.Conn())
	return client.GetNote(ctx, in, opts...)
}

// 判断笔记是否存在
func (m *defaultNote) IsNoteExist(ctx context.Context, in *IsNoteExistReq, opts ...grpc.CallOption) (*IsNoteExistRes, error) {
	client := NewNoteClient(m.cli.Conn())
	return client.IsNoteExist(ctx, in, opts...)
}
