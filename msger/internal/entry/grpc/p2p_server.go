package grpc

import (
	"context"
	"math"

	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
	p2pv1 "github.com/ryanreadbooks/whimer/msger/api/p2p/v1"
	bizp2p "github.com/ryanreadbooks/whimer/msger/internal/biz/p2p"
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

func (s *ChatServiceServer) SendMessage(ctx context.Context, in *p2pv1.SendMessageRequest) (
	*p2pv1.SendMessageResponse, error) {

	if in.Sender == 0 {
		return nil, global.ErrP2PChatSenderEmpty
	}
	if in.Receiver == 0 {
		return nil, global.ErrP2PChatReceiverEmpty
	}
	if in.ChatId <= 0 {
		return nil, global.ErrP2PChatNotExist
	}
	if in.Msg == nil {
		return nil, global.ErrChatMsgNil
	}
	if in.Msg.Type == pbmsg.MsgType_MSG_TYPE_UNSPECIFIED {
		return nil, global.ErrArgs.Msg("不支持的消息类型")
	}
	if len(in.Msg.Data) == 0 {
		return nil, global.ErrArgs.Msg("消息内容为空")
	}

	respMsg, err := s.Svc.P2PChatSrv.SendMessage(ctx, &bizp2p.CreateMsgReq{
		ChatId:   in.ChatId,
		Sender:   in.Sender,
		Receiver: in.Receiver,
		MsgType:  in.Msg.Type,
		Content:  in.Msg.Data,
	})
	if err != nil {
		return nil, err
	}

	return &p2pv1.SendMessageResponse{
		MsgId: respMsg.MsgId,
		Seq:   respMsg.Seq,
	}, nil
}

func (s *ChatServiceServer) ListMessage(ctx context.Context, in *p2pv1.ListMessageRequest) (
	*p2pv1.ListMessageResponse, error) {
	if err := checkChatIdUserId(in); err != nil {
		return nil, err
	}
	if in.Seq <= 0 {
		in.Seq = math.MaxInt64
	}
	if in.Count > 50 {
		in.Count = 50
	}

	msgs, nextSeq, err := s.Svc.P2PChatSrv.ListMessage(ctx, in.UserId, in.ChatId, in.Seq, in.Count)
	if err != nil {
		return nil, err
	}

	respMsgs := make([]*pbmsg.Message, 0, len(msgs))
	for _, m := range msgs {
		respMsgs = append(respMsgs, &pbmsg.Message{
			MsgId:    m.MsgId,
			ChatId:   m.ChatId,
			Sender:   m.Sender,
			Receiver: m.Receiver,
			Status:   m.Status,
			Content: &pbmsg.MsgContent{
				Type: m.Type,
				Data: m.Content,
			},
			Seq: m.Seq,
		})
	}

	return &p2pv1.ListMessageResponse{
		Messages: respMsgs,
		NextSeq:  nextSeq,
	}, nil
}

// 获取用户会话未读数
func (s *ChatServiceServer) GetUnreadCount(ctx context.Context, in *p2pv1.GetUnreadCountRequest) (
	*p2pv1.GetUnreadCountResponse, error) {
	if err := checkChatIdUserId(in); err != nil {
		return nil, err
	}

	cnt, err := s.Svc.P2PChatSrv.GetUnread(ctx, in.UserId, in.ChatId)
	if err != nil {
		return nil, err
	}

	return &p2pv1.GetUnreadCountResponse{
		Count: cnt,
	}, nil
}

// 清除未读数
func (s *ChatServiceServer) ClearUnread(ctx context.Context, in *p2pv1.ClearUnreadRequest) (
	*p2pv1.ClearUnreadResponse, error) {

	if err := checkChatIdUserId(in); err != nil {
		return nil, err
	}

	err := s.Svc.P2PChatSrv.ClearUnread(ctx, in.UserId, in.ChatId)
	if err != nil {
		return nil, err
	}

	return &p2pv1.ClearUnreadResponse{}, nil
}

// 撤回消息
func (s *ChatServiceServer) RevokeMessage(ctx context.Context, in *p2pv1.RevokeMessageRequest) (
	*p2pv1.RevokeMessageResponse, error) {

	if err := checkChatIdMsgId(in); err != nil {
		return nil, err
	}

	err := s.Svc.P2PChatSrv.RevokeMessage(ctx, in.ChatId, in.MsgId)
	if err != nil {
		return nil, err
	}

	return &p2pv1.RevokeMessageResponse{}, nil
}

type ChatIdUserIdGetter interface {
	GetChatId() int64
	GetUserId() int64
}

func checkChatIdUserId(g ChatIdUserIdGetter) error {
	if g.GetChatId() <= 0 {
		return global.ErrP2PChatNotExist
	}

	if g.GetUserId() == 0 {
		return global.ErrP2PChatUserEmpty
	}

	return nil
}

type ChatIdMsgIdGetter interface {
	GetChatId() int64
	GetMsgId() int64
}

func checkChatIdMsgId(g ChatIdMsgIdGetter) error {
	if g.GetChatId() <= 0 {
		return global.ErrP2PChatNotExist
	}

	if g.GetMsgId() == 0 {
		return global.ErrMsgNotExist
	}

	return nil
}
