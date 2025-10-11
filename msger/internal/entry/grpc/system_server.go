package grpc

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
	systemv1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
	bizsyschat "github.com/ryanreadbooks/whimer/msger/internal/biz/system"
	"github.com/ryanreadbooks/whimer/msger/internal/global/model"
	"github.com/ryanreadbooks/whimer/msger/internal/srv"
)

type SystemNotificationService struct {
	systemv1.UnimplementedNotificationServiceServer

	Svc *srv.Service
}

func NewSystemNotificationServiceServer(svc *srv.Service) *SystemNotificationService {
	return &SystemNotificationService{
		Svc: svc,
	}
}

// 系统通知
func (s *SystemNotificationService) NotifySystemNotice(ctx context.Context, in *systemv1.NotifySystemNoticeRequest) (
	*systemv1.NotifySystemNoticeResponse, error) {
	return nil, nil
}

// 回复我的
func (s *SystemNotificationService) NotifyReplyMsg(ctx context.Context, in *systemv1.NotifyReplyMsgRequest) (
	*systemv1.NotifyReplyMsgResponse, error) {
	return nil, nil
}

// @我的
func (s *SystemNotificationService) NotifyMentionMsg(ctx context.Context, in *systemv1.NotifyMentionMsgRequest) (
	*systemv1.NotifyMentionMsgResponse, error) {
	if len(in.Mentions) == 0 {
		return &systemv1.NotifyMentionMsgResponse{}, nil
	}

	// filter valid mentions
	reqs := make([]*model.SystemNotifyMentionMsg, 0, len(in.Mentions))
	for _, mentionReq := range in.Mentions {
		if !isMentionValid(mentionReq) {
			continue
		}

		reqs = append(reqs, &model.SystemNotifyMentionMsg{
			Uid:     mentionReq.GetUid(),
			Target:  mentionReq.GetTargetUid(),
			Content: mentionReq.GetContent(),
		})
	}

	msgIds, err := s.Svc.SystemChatSrv.NotifyMentionSystemMsg(ctx, reqs)
	if err != nil {
		return nil, err
	}

	respMsgIds := make(map[int64]*pbmsg.StringList)
	for recvUid, msgIds := range msgIds {
		respMsgIds[recvUid] = &pbmsg.StringList{
			Items: msgIds,
		}
	}

	return &systemv1.NotifyMentionMsgResponse{
		MsgIds: respMsgIds,
	}, nil
}

// 收到的赞
func (s *SystemNotificationService) NotifyLikesMsg(ctx context.Context, in *systemv1.NotifyLikesMsgRequest) (
	*systemv1.NotifyLikesMsgResponse, error) {
	return nil, nil
}

func handleSystemMsgCount(count int32) int32 {
	if count <= 0 {
		count = 20 // 默认值
	}
	if count > 100 {
		count = 100 // 最大值限制
	}
	return count
}

// 分页获取系统通知消息
func (s *SystemNotificationService) ListSystemNotifyMsg(ctx context.Context, in *systemv1.ListSystemNotifyMsgRequest) (
	*systemv1.ListSystemMsgResponse, error) {
	// 参数校验
	if in.GetRecvUid() == 0 {
		return nil, xerror.ErrArgs.Msg("invalid uid")
	}

	// 处理 count 参数
	count := handleSystemMsgCount(in.GetCount())

	msgs, hasMore, err := s.Svc.SystemChatSrv.ListSystemMsg(ctx, in.GetRecvUid(),
		model.SystemNotifyNoticeChat, in.GetCursor(), count)
	if err != nil {
		return nil, err
	}

	return &systemv1.ListSystemMsgResponse{
		Messages: convertSystemMsgsToResponse(msgs),
		HasMore:  hasMore,
	}, nil
}

// 分页获取系统回复消息
func (s *SystemNotificationService) ListSystemReplyMsg(ctx context.Context, in *systemv1.ListSystemReplyMsgRequest) (
	*systemv1.ListSystemMsgResponse, error) {
	// 参数校验
	if in.GetRecvUid() == 0 {
		return nil, xerror.ErrArgs.Msg("invalid uid")
	}

	// 处理 count 参数
	count := handleSystemMsgCount(in.GetCount())

	msgs, hasMore, err := s.Svc.SystemChatSrv.ListSystemMsg(ctx, in.GetRecvUid(),
		model.SystemNotifyReplyChat, in.GetCursor(), count)
	if err != nil {
		return nil, err
	}

	return &systemv1.ListSystemMsgResponse{
		Messages: convertSystemMsgsToResponse(msgs),
		HasMore:  hasMore,
	}, nil
}

// 分页获取系统@消息
func (s *SystemNotificationService) ListSystemMentionMsg(ctx context.Context, in *systemv1.ListSystemMentionMsgRequest) (
	*systemv1.ListSystemMsgResponse, error) {
	// 参数校验
	if in.GetRecvUid() == 0 {
		return nil, xerror.ErrArgs.Msg("invalid uid")
	}

	// 处理 count 参数
	count := handleSystemMsgCount(in.GetCount())

	msgs, hasMore, err := s.Svc.SystemChatSrv.ListSystemMsg(ctx, in.GetRecvUid(),
		model.SystemNotifyMentionedChat, in.GetCursor(), count)
	if err != nil {
		return nil, err
	}

	return &systemv1.ListSystemMsgResponse{
		Messages: convertSystemMsgsToResponse(msgs),
		HasMore:  hasMore,
	}, nil
}

// 分页获取系统点赞消息
func (s *SystemNotificationService) ListSystemLikesMsg(ctx context.Context, in *systemv1.ListSystemLikesMsgRequest) (
	*systemv1.ListSystemMsgResponse, error) {
	// 参数校验
	if in.GetRecvUid() == 0 {
		return nil, xerror.ErrArgs.Msg("invalid uid")
	}

	// 处理 count 参数
	count := handleSystemMsgCount(in.GetCount())

	// 调用 srv 层方法，传入点赞类型
	msgs, hasMore, err := s.Svc.SystemChatSrv.ListSystemMsg(ctx, in.GetRecvUid(),
		model.SystemNotifyLikesChat, in.GetCursor(), count)
	if err != nil {
		return nil, err
	}

	return &systemv1.ListSystemMsgResponse{
		Messages: convertSystemMsgsToResponse(msgs),
		HasMore:  hasMore,
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
