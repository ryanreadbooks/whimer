package grpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	systemv1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
	bizsyschat "github.com/ryanreadbooks/whimer/msger/internal/biz/system"
	"github.com/ryanreadbooks/whimer/msger/internal/global/model"
	"github.com/ryanreadbooks/whimer/msger/internal/srv"
)

type SystemChatServiceServer struct {
	systemv1.UnimplementedChatServiceServer

	Svc *srv.Service
}

func NewSystemChatServiceServer(svc *srv.Service) *SystemChatServiceServer {
	return &SystemChatServiceServer{
		Svc: svc,
	}
}

// 分页获取系统通知消息
func (s *SystemChatServiceServer) ListSystemNotifyMsg(ctx context.Context, in *systemv1.ListSystemNotifyMsgRequest) (
	*systemv1.ListSystemMsgResponse, error) {
	// 参数校验
	if in.GetRecvUid() == 0 {
		return nil, xerror.ErrArgs.Msg("invalid uid")
	}

	// 处理 count 参数
	count := handleSystemMsgCount(in.GetCount())

	resp, err := s.Svc.SystemChatSrv.ListSystemMsg(ctx, in.GetRecvUid(),
		model.SystemNotifyNoticeChat, in.GetCursor(), count)
	if err != nil {
		return nil, err
	}

	return &systemv1.ListSystemMsgResponse{
		Messages: convertSystemMsgsToResponse(resp.Msgs),
		HasMore:  resp.HasMore,
		ChatId:   convertChatIdToString(resp.ChatId),
	}, nil
}

// 分页获取系统回复消息
func (s *SystemChatServiceServer) ListSystemReplyMsg(ctx context.Context, in *systemv1.ListSystemReplyMsgRequest) (
	*systemv1.ListSystemMsgResponse, error) {
	// 参数校验
	if in.GetRecvUid() == 0 {
		return nil, xerror.ErrArgs.Msg("invalid uid")
	}

	// 处理 count 参数
	count := handleSystemMsgCount(in.GetCount())

	resp, err := s.Svc.SystemChatSrv.ListSystemMsg(ctx, in.GetRecvUid(),
		model.SystemNotifyReplyChat, in.GetCursor(), count)
	if err != nil {
		return nil, err
	}

	return &systemv1.ListSystemMsgResponse{
		Messages: convertSystemMsgsToResponse(resp.Msgs),
		HasMore:  resp.HasMore,
		ChatId:   convertChatIdToString(resp.ChatId),
	}, nil
}

// 分页获取系统@消息
func (s *SystemChatServiceServer) ListSystemMentionMsg(ctx context.Context, in *systemv1.ListSystemMentionMsgRequest) (
	*systemv1.ListSystemMsgResponse, error) {
	// 参数校验
	if in.GetRecvUid() == 0 {
		return nil, xerror.ErrArgs.Msg("invalid uid")
	}

	// 处理 count 参数
	count := handleSystemMsgCount(in.GetCount())

	resp, err := s.Svc.SystemChatSrv.ListSystemMsg(ctx, in.GetRecvUid(),
		model.SystemNotifyMentionedChat, in.GetCursor(), count)
	if err != nil {
		return nil, err
	}

	return &systemv1.ListSystemMsgResponse{
		Messages: convertSystemMsgsToResponse(resp.Msgs),
		HasMore:  resp.HasMore,
		ChatId:   convertChatIdToString(resp.ChatId),
	}, nil
}

// 分页获取系统点赞消息
func (s *SystemChatServiceServer) ListSystemLikesMsg(ctx context.Context, in *systemv1.ListSystemLikesMsgRequest) (
	*systemv1.ListSystemMsgResponse, error) {
	// 参数校验
	if in.GetRecvUid() == 0 {
		return nil, xerror.ErrArgs.Msg("invalid uid")
	}

	// 处理 count 参数
	count := handleSystemMsgCount(in.GetCount())

	// 调用 srv 层方法，传入点赞类型
	resp, err := s.Svc.SystemChatSrv.ListSystemMsg(ctx, in.GetRecvUid(),
		model.SystemNotifyLikesChat, in.GetCursor(), count)
	if err != nil {
		return nil, err
	}

	return &systemv1.ListSystemMsgResponse{
		Messages: convertSystemMsgsToResponse(resp.Msgs),
		HasMore:  resp.HasMore,
		ChatId:   convertChatIdToString(resp.ChatId),
	}, nil
}

// 将系统消息转换为响应格式
func convertSystemMsgsToResponse(msgs []*bizsyschat.SystemMsg) []*systemv1.SystemMsg {
	respMsgs := make([]*systemv1.SystemMsg, 0, len(msgs))
	for _, msg := range msgs {
		respMsgs = append(respMsgs, &systemv1.SystemMsg{
			Id:           msg.Id.String(),
			SystemChatId: msg.SystemChatId.String(),
			TriggerUid:   msg.TriggerUid,
			RecvUid:      msg.RecvUid,
			Status:       systemv1.SystemMsgStatus(msg.Status),
			MsgType:      msg.MsgType,
			Content:      msg.Content,
			Mtime:        msg.Mtime,
		})
	}

	return respMsgs
}

func convertChatIdToString(chatId uuid.UUID) string {
	if chatId.IsZero() {
		return ""
	}

	return chatId.String()
}

func isMentionValid(mention *systemv1.MentionMsgContent) bool {
	if mention.GetUid() == 0 || mention.GetTargetUid() == 0 || len(mention.GetContent()) == 0 {
		return false
	}

	// 排除自己@自己
	if mention.GetUid() == mention.GetTargetUid() {
		return false
	}

	return true
}

// 清除未读
func (s *SystemChatServiceServer) ClearChatUnread(ctx context.Context, in *systemv1.ClearChatUnreadRequest) (
	*systemv1.ClearChatUnreadResponse, error) {
	err := s.Svc.SystemChatSrv.ClearChatUnread(ctx, in.Uid, in.ChatId)

	return &systemv1.ClearChatUnreadResponse{}, err
}

func bizUnreadToPbUnread(u *bizsyschat.ChatUnread) *systemv1.ChatUnread {
	return &systemv1.ChatUnread{
		ChatId:      convertChatIdToString(u.ChatId),
		ChatType:    u.ChatType.Tag().String(),
		UnreadCount: u.UnreadCount,
	}
}

func bizUnreadsToPbUnreads(us []*bizsyschat.ChatUnread) []*systemv1.ChatUnread {
	resp := make([]*systemv1.ChatUnread, 0, len(us))
	for _, u := range us {
		resp = append(resp, bizUnreadToPbUnread(u))
	}

	return resp
}

// 获取单个chat的未读数
func (s *SystemChatServiceServer) GetChatUnread(ctx context.Context, in *systemv1.GetChatUnreadRequest) (
	*systemv1.GetChatUnreadResponse, error) {
	unread, err := s.Svc.SystemChatSrv.GetChatUnreadCount(ctx, in.Uid, in.ChatId)
	if err != nil {
		return nil, err
	}

	return &systemv1.GetChatUnreadResponse{
		Unread: bizUnreadToPbUnread(unread),
	}, nil
}

// 获取全部系统会话的未读数
func (s *SystemChatServiceServer) GetAllChatsUnread(ctx context.Context, in *systemv1.GetAllChatsUnreadRequest) (
	*systemv1.GetAllChatsUnreadResponse, error) {
	resp := systemv1.GetAllChatsUnreadResponse{}

	if in.Uid == 0 {
		return &resp, nil
	}

	unreads, err := s.Svc.SystemChatSrv.GetUserChatsUnreadCount(ctx, in.Uid)
	if err != nil {
		return nil, err
	}

	for _, ur := range unreads {
		switch ur.ChatType {
		case model.SystemNotifyNoticeChat:
			resp.NoticeUnread = bizUnreadToPbUnread(ur)
		case model.SystemNotifyReplyChat:
			resp.ReplyUnread = bizUnreadToPbUnread(ur)
		case model.SystemNotifyMentionedChat:
			resp.MentionUnread = bizUnreadToPbUnread(ur)
		case model.SystemNotifyLikesChat:
			resp.LikesUnread = bizUnreadToPbUnread(ur)
		}
	}

	return &resp, nil
}
