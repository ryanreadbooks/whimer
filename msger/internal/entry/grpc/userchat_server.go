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

func checkSendMsgToChatReq(msgType model.MsgType, in *pbuserchat.SendMsgToChatRequest) (u uuid.UUID, err error) {
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

	msgContent := in.Msg.GetContent()
	switch msgType {
	case model.MsgText:
		text, ok := msgContent.(*pbuserchat.MsgReq_Text)
		if !ok || text == nil {
			err = global.ErrUnsupportedMsgType
			return
		}
		if len(text.Text.GetContent()) == 0 {
			err = global.ErrArgs.Msg("content is empty for text msg")
			return
		}
	case model.MsgImage:
		image, ok := msgContent.(*pbuserchat.MsgReq_Image)
		if !ok || image == nil {
			err = global.ErrUnsupportedMsgType
			return
		}
		if err = model.CheckImageFormat(image.Image.Format); err != nil {
			return
		}
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

	msgType, err := model.MsgTypeFromPb(in.GetMsg().GetType())
	if err != nil {
		return nil, err
	}

	chatId, err := checkSendMsgToChatReq(msgType, in)
	if err != nil {
		return nil, err
	}
	req := &srv.SendMsgReq{
		Type: msgType,
		Cid:  in.Msg.Cid,
	}
	assignSendMsgReqContent(msgType, req, in)
	msgId, err := s.Srv.UserChatSrv.SendMsg(ctx, in.Sender, chatId, req)
	if err != nil {
		return nil, err
	}

	return &pbuserchat.SendMsgToChatResponse{MsgId: msgId.String()}, nil
}

func assignSendMsgReqContent(msgType model.MsgType, req *srv.SendMsgReq, pbIn *pbuserchat.SendMsgToChatRequest) {
	switch msgType {
	case model.MsgText:
		req.Text = ToBizMsgContentText(pbIn.Msg.Content.(*pbuserchat.MsgReq_Text))
	case model.MsgImage:
		req.Image = ToBizMsgContentImage(pbIn.Msg.Content.(*pbuserchat.MsgReq_Image))
	}
}

func (s *UserChatServiceServer) GetChatMembers(ctx context.Context, in *pbuserchat.GetChatMembersRequest) (
	*pbuserchat.GetChatMembersResponse, error) {
	chatId, err := uuid.ParseString(in.ChatId)
	if err != nil {
		return nil, global.ErrArgs.Msg("invalid chatid")
	}

	s.Srv.UserChatSrv.GetChatMembers(ctx, chatId)

	return &pbuserchat.GetChatMembersResponse{}, nil
}
