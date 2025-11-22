package grpc

import (
	"context"

	pbmsg "github.com/ryanreadbooks/whimer/msger/api/msg"
	systemv1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"

	"github.com/ryanreadbooks/whimer/msger/internal/model"
	"github.com/ryanreadbooks/whimer/msger/internal/srv"
)

type SystemNotificationServiceServer struct {
	systemv1.UnimplementedNotificationServiceServer

	Service *srv.Service
}

func NewSystemNotificationServiceServer(srv *srv.Service) *SystemNotificationServiceServer {
	return &SystemNotificationServiceServer{
		Service: srv,
	}
}

// 系统通知
func (s *SystemNotificationServiceServer) NotifySystemNotice(ctx context.Context, in *systemv1.NotifySystemNoticeRequest) (
	*systemv1.NotifySystemNoticeResponse, error) {
	return nil, nil
}

// 回复我的
func (s *SystemNotificationServiceServer) NotifyReplyMsg(ctx context.Context, in *systemv1.NotifyReplyMsgRequest) (
	*systemv1.NotifyReplyMsgResponse, error) {
	if len(in.GetContents()) == 0 {
		return &systemv1.NotifyReplyMsgResponse{}, nil
	}

	// filter invalid contents
	reqs := make([]*model.SystemNotifyReplyMsg, 0, len(in.Contents))
	for _, c := range in.Contents {
		if !isSysMsgContentValid(c) {
			continue
		}

		reqs = append(reqs, &model.SystemNotifyReplyMsg{
			Uid:     c.Uid,
			Target:  c.TargetUid,
			Content: c.Content,
		})
	}

	msgIds, err := s.Service.SystemChatSrv.NotifyReplySystemMsg(ctx, reqs)
	if err != nil {
		return nil, err
	}

	respMsgIds := make(map[int64]*pbmsg.StringList)
	for recvUid, msgIds := range msgIds {
		respMsgIds[recvUid] = &pbmsg.StringList{
			Items: msgIds,
		}
	}

	return &systemv1.NotifyReplyMsgResponse{MsgIds: respMsgIds}, nil
}

// @我的
func (s *SystemNotificationServiceServer) NotifyMentionMsg(ctx context.Context, in *systemv1.NotifyMentionMsgRequest) (
	*systemv1.NotifyMentionMsgResponse, error) {
	if len(in.Mentions) == 0 {
		return &systemv1.NotifyMentionMsgResponse{}, nil
	}

	// filter valid mentions
	reqs := make([]*model.SystemNotifyMentionMsg, 0, len(in.Mentions))
	for _, mentionReq := range in.Mentions {
		if !isSysMsgContentValid(mentionReq) {
			continue
		}

		reqs = append(reqs, &model.SystemNotifyMentionMsg{
			Uid:     mentionReq.GetUid(),
			Target:  mentionReq.GetTargetUid(),
			Content: mentionReq.GetContent(),
		})
	}

	msgIds, err := s.Service.SystemChatSrv.NotifyMentionSystemMsg(ctx, reqs)
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
func (s *SystemNotificationServiceServer) NotifyLikesMsg(ctx context.Context, in *systemv1.NotifyLikesMsgRequest) (
	*systemv1.NotifyLikesMsgResponse, error) {
	if len(in.Contents) == 0 {
		return &systemv1.NotifyLikesMsgResponse{}, nil
	}

	// filter valid mentions
	reqs := make([]*model.SystemNotifyLikesMsg, 0, len(in.Contents))
	for _, req := range in.Contents {
		if !isSysMsgContentValid(req) {
			continue
		}

		reqs = append(reqs, &model.SystemNotifyLikesMsg{
			Uid:     req.GetUid(),
			Target:  req.GetTargetUid(),
			Content: req.GetContent(),
		})
	}

	msgIds, err := s.Service.SystemChatSrv.NotifyLikesSystemMsg(ctx, reqs)
	if err != nil {
		return nil, err
	}

	respMsgIds := make(map[int64]*pbmsg.StringList)
	for recvUid, msgIds := range msgIds {
		respMsgIds[recvUid] = &pbmsg.StringList{
			Items: msgIds,
		}
	}

	return &systemv1.NotifyLikesMsgResponse{
		MsgIds: respMsgIds,
	}, nil
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
