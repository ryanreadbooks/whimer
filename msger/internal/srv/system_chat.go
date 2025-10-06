package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/msger/internal/biz"
	bizsyschat "github.com/ryanreadbooks/whimer/msger/internal/biz/system"
	"github.com/ryanreadbooks/whimer/msger/internal/global/model"
)

type SystemChatSrv struct {
	chatBiz bizsyschat.ChatBiz
}

func NewSystemChatSrv(biz biz.Biz) *SystemChatSrv {
	return &SystemChatSrv{
		chatBiz: biz.SystemBiz,
	}
}

func (s *SystemChatSrv) SendMsg(ctx context.Context, req bizsyschat.CreateSystemMsgReq) (uuid.UUID, error) {
	// TODO uid 系统通知设置检查
	msgId, err := s.chatBiz.CreateMsg(ctx, &req)
	if err != nil {
		return uuid.UUID{}, xerror.Wrapf(err, "srv failed to send system msg").WithCtx(ctx)
	}

	// TODO 消息推送

	return msgId.Id, nil
}

type NotifySystemMsgReq struct {
	RecvUid int64
	MsgType model.MsgType
	Notice  SystemNotice
}

type SystemNotice struct {
	Title   string
	Content string
}

// TODO 系统通知格式化
func (s *SystemNotice) Format() string {
	return s.Title + "\n" + s.Content
}

// 发送通用系统消息
func (s *SystemChatSrv) NotifySystemNoticeMsg(ctx context.Context, req NotifySystemMsgReq) (uuid.UUID, error) {
	return s.SendMsg(ctx, bizsyschat.CreateSystemMsgReq{
		TriggerUid: -1, // 系统
		RecvUid:    req.RecvUid,
		ChatType:   model.SystemNotifyNoticeChat,
		MsgType:    req.MsgType,
		Content:    req.Notice.Format(),
	})
}

type NotifyReplySystemMsgReq struct {
	ReplySender   int64
	ReplyReceiver int64
	MsgType       model.MsgType
	MsgContent    string
}

// 发送回复我的系统消息
func (s *SystemChatSrv) NotifyReplySystemMsg(ctx context.Context, req NotifyReplySystemMsgReq) (uuid.UUID, error) {
	return s.SendMsg(ctx, bizsyschat.CreateSystemMsgReq{
		TriggerUid: req.ReplySender, // 系统
		RecvUid:    req.ReplyReceiver,
		ChatType:   model.SystemNotifyReplyChat,
		MsgType:    req.MsgType,
		Content:    req.MsgContent,
	})
}

type NotifyMentionSystemMsgReq struct {
	MentionSender   int64
	MentionReceiver int64
	MsgType         model.MsgType
	MsgContent      string
}

// 发送@我的系统消息
func (s *SystemChatSrv) NotifyMentionSystemMsg(ctx context.Context, req NotifyMentionSystemMsgReq) (uuid.UUID, error) {
	return s.SendMsg(ctx, bizsyschat.CreateSystemMsgReq{
		TriggerUid: req.MentionSender, // 系统
		RecvUid:    req.MentionReceiver,
		ChatType:   model.SystemNotifyMentionedChat,
		MsgType:    req.MsgType,
		Content:    req.MsgContent,
	})
}

type NotifyLikesSystemMsgReq struct {
	LikesSender   int64
	LikesReceiver int64
	MsgType       model.MsgType
	MsgContent    string
}

// 发送收到的赞系统消息
func (s *SystemChatSrv) NotifyLikesSystemMsg(ctx context.Context, req NotifyLikesSystemMsgReq) (uuid.UUID, error) {
	return s.SendMsg(ctx, bizsyschat.CreateSystemMsgReq{
		TriggerUid: req.LikesSender, // 系统
		RecvUid:    req.LikesReceiver,
		ChatType:   model.SystemNotifyLikesChat,
		MsgType:    req.MsgType,
		Content:    req.MsgContent,
	})
}
