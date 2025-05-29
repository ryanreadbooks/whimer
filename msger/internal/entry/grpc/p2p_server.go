package grpc

import (
	"context"

	p2pv1 "github.com/ryanreadbooks/whimer/msger/api/p2p/v1"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/srv"

	"github.com/bufbuild/protovalidate-go"
)

type ChatServiceServer struct {
	p2pv1.UnimplementedChatServiceServer
	validator *protovalidate.Validator

	Svc *srv.Service
}

func NewChatServiceServer(svc *srv.Service) *ChatServiceServer {
	return &ChatServiceServer{
		Svc: svc,
	}
}

func (s *ChatServiceServer) CreateChat(ctx context.Context, in *p2pv1.CreateChatRequest) (
	*p2pv1.CreateChatResponse, error) {
	if in.Initiator == 0 || in.Target == 0 {
		return nil, global.ErrP2PChatUserEmpty
	}

	chatId, err := s.Svc.P2PChatSrv.CreateChat(ctx, in.Initiator, in.Target)
	if err != nil {
		return nil, err
	}

	return &p2pv1.CreateChatResponse{
		ChatId: chatId,
	}, nil
}
