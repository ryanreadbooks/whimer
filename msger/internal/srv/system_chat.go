package srv

import (
	"context"
	"sort"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/msger/internal/biz"
	bizsyschat "github.com/ryanreadbooks/whimer/msger/internal/biz/system"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/model"

	"golang.org/x/sync/errgroup"
)

type SystemChatSrv struct {
	chatBiz bizsyschat.ChatBiz
}

func NewSystemChatSrv(biz biz.Biz) *SystemChatSrv {
	return &SystemChatSrv{
		chatBiz: biz.SystemBiz,
	}
}

// 单个用户发送通知（消息落库）
func (s *SystemChatSrv) saveMsg(ctx context.Context, req *bizsyschat.CreateSystemMsgReq) (uuid.UUID, error) {
	// TODO uid 系统通知设置检查

	msgId, err := s.chatBiz.CreateMsg(ctx, req)
	if err != nil {
		return uuid.UUID{}, xerror.Wrapf(err, "system chat srv failed to send system msg").WithCtx(ctx)
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

func (s *SystemChatSrv) notifySystemMsg(ctx context.Context,
	chatType model.SystemChatType, reqs []model.ISystemMsg) (map[int64][]string, error) {

	var msgReqs = make([]*bizsyschat.CreateSystemMsgReq, 0, len(reqs))
	for _, req := range reqs {
		msgReqs = append(msgReqs, &bizsyschat.CreateSystemMsgReq{
			TriggerUid: req.GetUid(),
			RecvUid:    req.GetTargetUid(),
			ChatType:   chatType,
			MsgType:    model.MsgText,
			Content:    req.GetContent(), // MentionMsgContent
		})
	}

	var targetMsgIds = make(map[int64][]string, len(msgReqs))
	var successTargets = make(map[int64][]*bizsyschat.CreateSystemMsgReq, len(msgReqs))

	if len(msgReqs) == 0 {
		return targetMsgIds, nil
	}

	wg := errgroup.Group{}
	// 记录创建消息成功的记录
	for _, req := range msgReqs {
		wg.Go(func() error {
			msgId, err := s.saveMsg(ctx, req)
			if err != nil {
				xlog.Msg("system chat srv failed to save mention system msg").
					Extras("uid", req.TriggerUid, "recv_uid", req.RecvUid).
					Err(err).Errorx(ctx)
				return nil // 不返回err
			}

			req.MsgId = msgId
			// 落库成功将msgId返回
			targetMsgIds[req.RecvUid] = append(targetMsgIds[req.RecvUid], msgId.String())
			successTargets[req.RecvUid] = append(successTargets[req.RecvUid], req)
			return nil
		})
	}

	if err := wg.Wait(); err != nil {
		return nil, xerror.Wrapf(err, "system chat srv failed to send mention system msg").WithCtx(ctx)
	}

	// 返回成成功创建的msgIds
	return targetMsgIds, nil
}

// 发送@我的系统消息
func (s *SystemChatSrv) NotifyMentionSystemMsg(ctx context.Context,
	reqs []*model.SystemNotifyMentionMsg) (map[int64][]string, error) {

	iMsgReqs := model.MakeSystemNotifyMentionMsgAsSlice(reqs)
	return s.notifySystemMsg(ctx, model.SystemNotifyMentionedChat, iMsgReqs)
}

// 发送回复我的系统消息
func (s *SystemChatSrv) NotifyReplySystemMsg(ctx context.Context,
	reqs []*model.SystemNotifyReplyMsg) (map[int64][]string, error) {

	iMsgReqs := model.MakeSystemNotifyReplyMsgAsSlice(reqs)
	return s.notifySystemMsg(ctx, model.SystemNotifyReplyChat, iMsgReqs)
}

// 发送收到的赞系统消息
func (s *SystemChatSrv) NotifyLikesSystemMsg(ctx context.Context,
	reqs []*model.SystemNotifyLikesMsg) (map[int64][]string, error) {

	iMsgReqs := model.MakeSystemNotifyLikesMsgAsSlice(reqs)
	return s.notifySystemMsg(ctx, model.SystemNotifyLikesChat, iMsgReqs)
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
		return resp, xerror.Wrapf(err, "system chat srv failed to list system msg").WithCtx(ctx)
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
		return xerror.Wrapf(err, "system chat srv failed to clear chat unread")
	}

	return nil
}

// 获取单个会话未读数
func (s *SystemChatSrv) GetChatUnreadCount(ctx context.Context, uid int64, chatId string) (*bizsyschat.ChatUnread, error) {
	cid, err := uuid.ParseString(chatId)
	if err != nil {
		return nil, global.ErrSysChatNotExist
	}

	unread, err := s.chatBiz.GetChatUnreadCount(ctx, uid, cid)
	if err != nil {
		return nil, xerror.Wrapf(err, "system chat srv failed to get chat unread count")
	}

	return unread, nil
}

// 获取某个用户所有系统会话的未读数
func (s *SystemChatSrv) GetUserChatsUnreadCount(ctx context.Context, uid int64) ([]*bizsyschat.ChatUnread, error) {
	unreads, err := s.chatBiz.GetUserAllChatUnreadCount(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "system chat srv failed to get user chat unread count")
	}

	// 补全缺失的chat
	for _, t := range model.SystemChatTypeSlice {
		found := false
		for _, u := range unreads {
			if u.ChatType == t {
				found = true
				break
			}
		}

		if !found {
			unreads = append(unreads, &bizsyschat.ChatUnread{
				ChatId:      uuid.EmptyUUID(),
				ChatType:    t,
				UnreadCount: 0,
			})
		}
	}

	sort.Slice(unreads, func(i, j int) bool { return unreads[i].ChatType < unreads[j].ChatType })

	return unreads, nil
}

// 删除系统会话消息
func (s *SystemChatSrv) DeleteSystemMsg(ctx context.Context, uid int64, msgId string) error {
	msgUUID, err := uuid.ParseString(msgId)
	if err != nil {
		return nil
	}

	chatId, err := s.chatBiz.GetMsgChatId(ctx, msgUUID)
	if err != nil {
		return xerror.Wrapf(err, "system chat srv failed to get chat_id by msg_id")
	}

	err = s.chatBiz.DeleteMsg(ctx, chatId, msgUUID, uid)
	if err != nil {
		return xerror.Wrapf(err, "system chat srv failed to delete msg")
	}

	return nil
}
