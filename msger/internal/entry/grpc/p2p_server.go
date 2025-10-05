package grpc

import (
	"context"
	"math"
	"unicode/utf8"

	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
	p2pv1 "github.com/ryanreadbooks/whimer/msger/api/p2p/v1"
	bizp2p "github.com/ryanreadbooks/whimer/msger/internal/biz/p2p"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/global/model"
	"github.com/ryanreadbooks/whimer/msger/internal/srv"
)

type ChatServiceServer struct {
	p2pv1.UnimplementedChatServiceServer

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

func validateMsgType(t pbmsg.MsgType) bool {
	return t == pbmsg.MsgType_MSG_TYPE_TEXT || t == pbmsg.MsgType_MSG_TYPE_IMAGE
}

func validateSendMsgRequest(in *p2pv1.SendMsgRequest) error {
	if in.Sender == 0 {
		return global.ErrP2PChatSenderEmpty
	}
	if in.Receiver == 0 {
		return global.ErrP2PChatReceiverEmpty
	}
	if in.ChatId <= 0 {
		return global.ErrP2PChatNotExist
	}
	if in.Msg == nil {
		return global.ErrChatMsgNil
	}
	if !validateMsgType(in.Msg.Type) {
		return global.ErrUnsupportedMsgType
	}

	dataLen := utf8.RuneCountInString(in.Msg.Data)
	if dataLen == 0 {
		return global.ErrEmptyMsg
	}
	if in.Msg.Type == pbmsg.MsgType_MSG_TYPE_TEXT {
		if dataLen > model.MaxTextLength {
			return global.ErrArgs.Msg("消息长度太长")
		}
	} else if in.Msg.Type == pbmsg.MsgType_MSG_TYPE_IMAGE {
		// check image is the right format
		// TODO 会话中的图片是否应该限制其它用户访问？
		return global.ErrUnsupportedMsgType
	}

	return nil
}

func (s *ChatServiceServer) SendMsg(ctx context.Context, in *p2pv1.SendMsgRequest) (
	*p2pv1.SendMsgResponse, error) {
	if err := validateSendMsgRequest(in); err != nil {
		return nil, err
	}

	respMsg, err := s.Svc.P2PChatSrv.SendMsg(ctx, &bizp2p.CreateMsgReq{
		ChatId:   in.ChatId,
		Sender:   in.Sender,
		Receiver: in.Receiver,
		MsgType:  in.Msg.Type,
		Content:  in.Msg.Data,
	})
	if err != nil {
		return nil, err
	}

	return &p2pv1.SendMsgResponse{
		MsgId: respMsg.MsgId,
		Seq:   respMsg.Seq,
	}, nil
}

func (s *ChatServiceServer) ListMsg(ctx context.Context, in *p2pv1.ListMsgRequest) (
	*p2pv1.ListMsgResponse, error) {
	if err := checkChatIdUserId(in); err != nil {
		return nil, err
	}
	if in.Seq <= 0 {
		in.Seq = math.MaxInt64
	}
	if in.Count > 50 {
		in.Count = 50
	}

	msgs, nextSeq, err := s.Svc.P2PChatSrv.ListMsg(ctx, in.UserId, in.ChatId, in.Seq, in.Count)
	if err != nil {
		return nil, err
	}

	respMsgs := make([]*pbmsg.Msg, 0, len(msgs))
	for _, m := range msgs {
		respMsgs = append(respMsgs, makePbMsg(m))
	}

	return &p2pv1.ListMsgResponse{
		Msgs: respMsgs,
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
func (s *ChatServiceServer) RevokeMsg(ctx context.Context, in *p2pv1.RevokeMsgRequest) (
	*p2pv1.RevokeMsgResponse, error) {
	if err := checkChatIdUserId(in); err != nil {
		return nil, err
	}
	if err := checkChatIdMsgId(in); err != nil {
		return nil, err
	}

	err := s.Svc.P2PChatSrv.RevokeMsg(ctx, in.UserId, in.ChatId, in.MsgId)
	if err != nil {
		return nil, err
	}

	return &p2pv1.RevokeMsgResponse{}, nil
}

func (s *ChatServiceServer) ListChat(ctx context.Context, in *p2pv1.ListChatRequest) (
	*p2pv1.ListChatResponse, error) {
	if in.UserId == 0 {
		return nil, global.ErrP2PChatUserEmpty
	}
	if in.Count > 50 {
		in.Count = 50
	}
	if in.Seq <= 0 {
		in.Seq = math.MaxInt64
	}

	chats, nextSeq, err := s.Svc.P2PChatSrv.ListChat(ctx, in.UserId, in.Seq, in.Count)
	if err != nil {
		return nil, err
	}

	result := make([]*p2pv1.Chat, 0, len(chats))
	for _, c := range chats {
		result = append(result, makeChatFromBiz(c))
	}

	return &p2pv1.ListChatResponse{
		Chats:   result,
		NextSeq: nextSeq,
	}, nil
}

func (s *ChatServiceServer) GetChat(ctx context.Context, in *p2pv1.GetChatRequest) (*p2pv1.GetChatResponse, error) {
	if in.ChatId == 0 {
		return nil, global.ErrP2PChatNotExist
	}

	chat, err := s.Svc.P2PChatSrv.GetChat(ctx, in.UserId, in.ChatId)
	if err != nil {
		return nil, err
	}

	return &p2pv1.GetChatResponse{
		Chat: makeChatFromBiz(chat),
	}, nil
}

func makeChatFromBiz(c *bizp2p.Chat) *p2pv1.Chat {
	return &p2pv1.Chat{
		ChatId:        c.ChatId,
		UserId:        c.UserId,
		PeerId:        c.PeerId,
		Unread:        c.Unread,
		LastMsgId:     c.LastMsgId,
		LastMsgSeq:    c.LastMsgSeq,
		LastReadMsgId: c.LastReadMsgId,
		LastReadTime:  c.LastReadTime,
		LastMsg:       makePbMsg(c.LastMsg),
	}
}

// 删除会话
func (s *ChatServiceServer) DeleteChat(ctx context.Context, in *p2pv1.DeleteChatRequest) (
	*p2pv1.DeleteChatResponse, error) {
	if err := checkChatIdUserId(in); err != nil {
		return nil, err
	}

	if err := s.Svc.P2PChatSrv.DeleteChat(ctx, in.UserId, in.ChatId); err != nil {
		return nil, err
	}

	return &p2pv1.DeleteChatResponse{}, nil
}

// 删除消息
func (s *ChatServiceServer) DeleteMsg(ctx context.Context, in *p2pv1.DeleteMsgRequest) (
	*p2pv1.DeleteMsgResponse, error) {
	if err := checkChatIdUserId(in); err != nil {
		return nil, err
	}
	if err := checkChatIdMsgId(in); err != nil {
		return nil, err
	}

	if err := s.Svc.P2PChatSrv.DeleteChatMsg(ctx, in.UserId, in.ChatId, in.MsgId); err != nil {
		return nil, err
	}

	return &p2pv1.DeleteMsgResponse{}, nil
}
