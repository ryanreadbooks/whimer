package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/msger/api/msg"
	"github.com/ryanreadbooks/whimer/msger/internal/biz"
	bizsyschat "github.com/ryanreadbooks/whimer/msger/internal/biz/system"
	bizwebsocket "github.com/ryanreadbooks/whimer/msger/internal/biz/websocket"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/global/model"

	"golang.org/x/sync/errgroup"
)

type SystemChatSrv struct {
	chatBiz      bizsyschat.ChatBiz
	websocketBiz bizwebsocket.Biz
}

func NewSystemChatSrv(biz biz.Biz) *SystemChatSrv {
	return &SystemChatSrv{
		chatBiz:      biz.SystemBiz,
		websocketBiz: biz.WebsocketBiz,
	}
}

// 单个用户发送通知（消息落库）
func (s *SystemChatSrv) saveMsg(ctx context.Context, req *bizsyschat.CreateSystemMsgReq) (uuid.UUID, error) {
	// TODO uid 系统通知设置检查

	msgId, err := s.chatBiz.CreateMsg(ctx, req)
	if err != nil {
		return uuid.UUID{}, xerror.Wrapf(err, "srv failed to send system msg").WithCtx(ctx)
	}

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
	return s.saveMsg(ctx, &bizsyschat.CreateSystemMsgReq{
		TriggerUid: -1, // 系统
		RecvUid:    req.RecvUid,
		ChatType:   model.SystemNotifyNoticeChat,
		MsgType:    req.MsgType,
	})
}

type NotifyReplySystemMsgReq struct {
	ReplySender   int64
	ReplyReceiver int64
	MsgType       model.MsgType
	MsgContent    []byte
}

// 发送回复我的系统消息
func (s *SystemChatSrv) NotifyReplySystemMsg(ctx context.Context, req NotifyReplySystemMsgReq) (uuid.UUID, error) {
	return s.saveMsg(ctx, &bizsyschat.CreateSystemMsgReq{
		TriggerUid: req.ReplySender, // 系统
		RecvUid:    req.ReplyReceiver,
		ChatType:   model.SystemNotifyReplyChat,
		MsgType:    req.MsgType,
		Content:    req.MsgContent,
	})
}

// 发送@我的系统消息
func (s *SystemChatSrv) NotifyMentionSystemMsg(ctx context.Context, reqs []*model.SystemNotifyMentionMsg) (map[int64][]string, error) {
	var msgReqs = make([]*bizsyschat.CreateSystemMsgReq, 0, len(reqs))
	for _, req := range reqs {
		msgReqs = append(msgReqs, &bizsyschat.CreateSystemMsgReq{
			TriggerUid: req.Uid,
			RecvUid:    req.Target,
			ChatType:   model.SystemNotifyMentionedChat,
			MsgType:    msg.MsgType_MSG_TYPE_TEXT,
			Content:    req.Content, // MentionMsgContent
		})
	}

	var targetMsgIds = make(map[int64][]string, len(msgReqs))
	var targetReqs = make(map[int64][]*bizsyschat.CreateSystemMsgReq, len(msgReqs))

	if len(msgReqs) == 0 {
		return targetMsgIds, nil
	}

	wg := errgroup.Group{}
	// 记录创建消息成功的记录
	for _, req := range msgReqs {
		req := req
		wg.Go(func() error {
			msgId, err := s.saveMsg(ctx, req)
			if err != nil {
				xlog.Msg("srv failed to save mention system msg").
					Extras("uid", req.TriggerUid, "recv_uid", req.RecvUid).
					Err(err).Errorx(ctx)
				return nil // 不返回err
			}

			req.MsgId = msgId
			// 落库成功将msgId返回
			targetMsgIds[req.RecvUid] = append(targetMsgIds[req.RecvUid], msgId.String())
			targetReqs[req.RecvUid] = append(targetReqs[req.RecvUid], req)
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, xerror.Wrapf(err, "srv failed to send mention system msg").WithCtx(ctx)
	}

	// 通知recvUid
	for recvUid, reqs := range targetReqs {
		targets := make([]*bizwebsocket.NotifySysContent, 0, len(reqs))
		for _, t := range targetReqs[recvUid] {
			webReq := &bizwebsocket.NotifySysContent{
				MsgId:   t.MsgId.String(),
				MsgType: t.MsgType,
				Content: t.Content, // MentionMsgContent
			}
			targets = append(targets, webReq)
		}

		if err := s.websocketBiz.NotifySysMention(ctx, recvUid, targets); err != nil {
			xlog.Msg("srv failed to notify sys mention").
				Extras("recv_uid", recvUid).
				Err(err).Errorx(ctx)
		}
	}

	return targetMsgIds, nil
}

type NotifyLikesSystemMsgReq struct {
	LikesSender   int64
	LikesReceiver int64
	MsgType       model.MsgType
	MsgContent    []byte
}

// 发送收到的赞系统消息
func (s *SystemChatSrv) NotifyLikesSystemMsg(ctx context.Context, req NotifyLikesSystemMsgReq) (uuid.UUID, error) {
	return s.saveMsg(ctx, &bizsyschat.CreateSystemMsgReq{
		TriggerUid: req.LikesSender, // 系统
		RecvUid:    req.LikesReceiver,
		ChatType:   model.SystemNotifyLikesChat,
		MsgType:    req.MsgType,
		Content:    req.MsgContent,
	})
}

// 分页获取系统消息
func (s *SystemChatSrv) ListSystemMsg(ctx context.Context,
	recvUid int64, chatType model.SystemChatType,
	cursor string, count int32) (*bizsyschat.ListMsgResp, error) {

	resp, err := s.chatBiz.ListMsg(ctx, &bizsyschat.ListMsgReq{
		RecvUid:  recvUid,
		ChatType: chatType,
		Cursor:   cursor,
		Count:    count,
	})
	if err != nil {
		return resp, xerror.Wrapf(err, "srv failed to list system msg").WithCtx(ctx)
	}

	return resp, nil
}

// 清除未读
func (s *SystemChatSrv) ClearChatUnread(ctx context.Context, uid int64, chatId string) error {
	cid, err := uuid.ParseString(chatId)
	if err != nil {
		return global.ErrSysChatNotExist
	}

	err = s.chatBiz.ClearChatUnread(ctx, uid, cid)
	if err != nil {
		return xerror.Wrapf(err, "srv failed to clear chat unread")
	}

	return nil
}
