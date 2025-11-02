package grpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	pbuserchat "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
	"github.com/ryanreadbooks/whimer/msger/internal/srv"
)

type UserChatServiceServer struct {
	pbuserchat.UnimplementedUserChatServiceServer

	Srv *srv.Service
}

func NewUserChatServiceServer(svc *srv.Service) *UserChatServiceServer {
	return &UserChatServiceServer{
		Srv: svc,
	}
}

func checkSendMsgToChatReq(in *pbuserchat.SendMsgToChatRequest) (u uuid.UUID, err error) {
	if in.Sender == 0 {
		err = global.ErrArgs.Msg("invalid sender")
		return
	}

	if len(in.ChatId) == 0 {
		err = global.ErrArgs.Msg("empty chatid")
		return
	}

	chatId, err := uuid.ParseString(in.ChatId)
	if err != nil {
		err = global.ErrArgs.Msg("invalid chatid")
		return
	}

	if in.Msg == nil {
		err = global.ErrArgs.Msg("msg req is nil")
		return
	}

	if len(in.Msg.GetContent()) == 0 {
		err = global.ErrArgs.Msg("empty msg content")
		return
	}

	return chatId, nil
}

// 发起单聊
func (s *UserChatServiceServer) CreateP2PChat(ctx context.Context,
	in *pbuserchat.CreateP2PChatRequest) (*pbuserchat.CreateP2PChatResponse, error) {

	initer := in.GetUid()
	target := in.GetTarget()

	chatId, err := s.Srv.UserChatSrv.InitP2PChat(ctx, initer, target)
	if err != nil {
		return nil, err
	}

	return &pbuserchat.CreateP2PChatResponse{ChatId: chatId.String()}, nil
}

// 在会话中发送消息
func (s *UserChatServiceServer) SendMsgToChat(ctx context.Context,
	in *pbuserchat.SendMsgToChatRequest) (*pbuserchat.SendMsgToChatResponse, error) {
	chatId, err := checkSendMsgToChatReq(in)
	if err != nil {
		return nil, err
	}

	msgType, err := model.MsgTypeFromPb(in.GetMsg().GetType())
	if err != nil {
		return nil, err
	}

	msgId, err := s.Srv.UserChatSrv.SendMsg(ctx, in.Sender, chatId, &srv.SendMsgReq{
		Type:    msgType,
		Content: in.Msg.Content,
		Cid:     in.Msg.Cid,
	})
	if err != nil {
		return nil, err
	}

	return &pbuserchat.SendMsgToChatResponse{MsgId: msgId.String()}, nil
}
