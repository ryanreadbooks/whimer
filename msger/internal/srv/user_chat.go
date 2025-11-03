package srv

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/msger/internal/biz"
	"github.com/ryanreadbooks/whimer/msger/internal/biz/userchat"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/infra"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
	"golang.org/x/sync/errgroup"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	emptyUUID = uuid.EmptyUUID()
)

type UserChatSrv struct {
	chatBiz       userchat.ChatBiz
	chatMemberBiz userchat.ChatMemberBiz
	msgBiz        userchat.MsgBiz
	chatInboxBiz  userchat.ChatInboxBiz
}

func NewUserChatSrv(biz biz.Biz) *UserChatSrv {
	return &UserChatSrv{
		chatBiz:       biz.ChatBiz,
		chatMemberBiz: biz.ChatMemberBiz,
		msgBiz:        biz.MsgBiz,
		chatInboxBiz:  biz.ChatInboxBiz,
	}
}

// 创建单聊
//
// initer: 发起单聊的用户uid;
// target: 目标单聊用户uid
func (s *UserChatSrv) InitP2PChat(ctx context.Context, initer, target int64) (uuid.UUID, error) {
	userA, userB := initer, target
	if userA > userB {
		userA, userB = userB, userA
	}

	lockKey := fmt.Sprintf("msger.userchat.lock.initp2pchat.%d:%d", userA, userB)
	locker := redis.NewRedisLock(infra.Redis(), lockKey)
	acquired, err := locker.AcquireCtx(ctx)
	if err != nil {
		return emptyUUID, xerror.Wrapf(err, "acquire lock failed").
			WithExtras("userA", userA, "userB", userB).WithCtx(ctx)
	}
	defer locker.ReleaseCtx(ctx)
	if !acquired {
		return emptyUUID, global.ErrLockNotHeld
	}

	chatId, err := s.chatMemberBiz.GetP2PChatId(ctx, userA, userB)
	if err != nil {
		if !errors.Is(err, global.ErrChatNotExist) {
			return emptyUUID, xerror.Wrapf(err, "member biz get p2p failed").WithCtx(ctx)
		}

		// p2p chat not exists
		err = infra.DaoTransact(ctx, func(ctx context.Context) error {
			// create p2p chat
			newChatId, err := s.chatBiz.CreateP2PChat(ctx)
			if err != nil {
				return xerror.Wrapf(err, "chat biz create p2p failed").WithCtx(ctx)
			}

			// create p2p members
			err = s.chatMemberBiz.InsertP2PMembers(ctx, newChatId, userA, userB)
			if err != nil {
				return xerror.Wrapf(err, "member biz insert p2p failed").WithCtx(ctx)
			}

			chatId = newChatId
			return nil
		})
		if err != nil {
			return emptyUUID, xerror.Wrapf(err, "dao transact failed").
				WithExtras("userA", userA, "userB", userB).WithCtx(ctx)
		}
	}

	// insert inbox for chat initer
	err = s.chatInboxBiz.PrepareInbox(ctx, initer, chatId)
	if err != nil {
		// 这里初始化信箱失败
		return emptyUUID, xerror.Wrapf(err, "chat inbox biz prepare failed").WithCtx(ctx)
	}

	return chatId, nil
}

// 往会话中发送消息

// 发送消息 主要步骤包含:
//
//  1. 创建消息
//  2. 消息绑定会话
//  3. 更新会话最后一条消息
//  4. 更新所有user的收件箱, 所有user的收件箱的插入新信件(更新lastMsgId)
func (s *UserChatSrv) SendMsg(ctx context.Context,
	sender int64, chatId uuid.UUID, msgReq *SendMsgReq) (uuid.UUID, error) {

	noMsgId := uuid.EmptyUUID()

	targetChat, err := s.chatBiz.GetChat(ctx, chatId)
	if err != nil {
		return noMsgId, xerror.Wrapf(err, "chat biz get chat failed").WithCtx(ctx)
	}

	err = s.chatMemberBiz.AttachChatMembers(ctx, targetChat)
	if err != nil {
		return noMsgId, xerror.Wrapf(err, "chat member biz attach chat failed").WithCtx(ctx)
	}

	// check if user can send to chat
	err = s.isAllowedToSendMsg(ctx, sender, targetChat)
	if err != nil {
		return noMsgId, xerror.Wrapf(err, "sender unable to send").
			WithExtras("sender", sender).WithCtx(ctx)
	}

	// TODO check chat status and settings
	if !targetChat.IsStatusNormal() {
		return noMsgId, xerror.Wrap(global.ErrChatNotNormal)
	}

	// 异步准备收件箱
	eg, gctx := errgroup.WithContext(ctx)
	eg.Go(recovery.DoV2(func() error {
		err := s.chatInboxBiz.BatchPrepareInboxes(gctx, chatId, targetChat.Members)
		if err != nil {
			return xerror.Wrapf(err, "async chat inbox biz batch prepare failed").WithCtx(gctx)
		}

		return nil
	}))

	var newMsg *userchat.Msg
	switch {
	case targetChat.IsP2PChat():
		newMsg, err = s.sendP2PMsg(ctx, sender, targetChat, msgReq, targetChat.Members)
	case targetChat.IsGroupChat():
		newMsg, err = s.sendGroupMsg(ctx, sender, targetChat, msgReq, targetChat.Members)
	default:
		return noMsgId, xerror.Wrap(global.ErrUnsupportedChatType)
	}

	if err != nil {
		return noMsgId, xerror.Wrapf(err, "user chat send msg failed").WithCtx(ctx)
	}

	gerr := eg.Wait()
	if gerr != nil {
		xlog.Msg("send msg async work failed").Err(err).Errorx(ctx)
		// do prepare again
		err := s.chatInboxBiz.BatchPrepareInboxes(gctx, chatId, targetChat.Members)
		if err != nil {
			return noMsgId, xerror.Wrapf(err, "chat inbox biz batch prepare failed").WithCtx(ctx)
		}
	}

	// 写入信箱
	err = s.updateUserInboxes(ctx, targetChat, targetChat.Members, newMsg)
	if err != nil {
		return noMsgId, xerror.Wrapf(err, "update user inboxes failed").WithCtx(ctx)
	}

	// 发送者的信箱需要更新最后已读
	err = s.chatInboxBiz.SetLastReadMsgIdToLatest(ctx, chatId, sender)
	if err != nil {
		// 仅打日志
		xlog.Msgf("chat inbox biz set last read msg id for %d failed", sender).
			Err(err).Errorx(ctx)
	}

	return newMsg.Id, nil
}

// operator撤回会话chatId中的消息msgId
func (s *UserChatSrv) RecallMsg(ctx context.Context, operator int64, chatId, msgId uuid.UUID) (err error) {
	logExtras := []any{
		"chat_id", chatId,
		"msg_id", msgId,
	}

	targetChat, err := s.chatBiz.GetChat(ctx, chatId)
	if err != nil {
		return xerror.Wrapf(err, "chat biz get chat failed").
			WithExtras(logExtras...).WithCtx(ctx)
	}

	err = s.chatMemberBiz.AttachChatMembers(ctx, targetChat)
	if err != nil {
		return xerror.Wrapf(err, "chat member biz attach failed").
			WithExtras(logExtras...).WithCtx(ctx)
	}

	targetMsg, err := s.msgBiz.GetMsg(ctx, msgId)
	if err != nil {
		return xerror.Wrapf(err, "msg biz get msg failed").
			WithExtras(logExtras...).WithCtx(ctx)
	}

	// check if operator can recall targetMsg
	err = s.isAllowedToRecallMsg(ctx, operator, targetChat, targetMsg)
	if err != nil {
		return xerror.Wrap(err)
	}

	err = infra.DaoTransact(ctx, func(ctx context.Context) error {
		err := s.msgBiz.RecallMsg(ctx, operator, targetMsg)
		if err != nil {
			return xerror.Wrapf(err, "msg biz recall msg failed").
				WithExtras(logExtras...).WithCtx(ctx)
		}
		return nil
	})
	if err != nil {
		return xerror.Wrapf(err, "tx recall failed")
	}

	// 处理inbox 如果msgId还没有被读 需要撤回未读数？
	if targetChat.IsP2PChat() {
		reads, err := s.BatchCheckUsersInBoxMsgRead(ctx, chatId, msgId, targetChat.Members)
		if err != nil {
			// 消息已经撤回 未读数为更新失败不影响流程
			xlog.Msgf("batch check users inbox msg read error").
				Extras("chat_id", chatId, "msg_id", msgId).Err(err).Errorx(ctx)
			return nil
		}
		uidA, uidB := targetChat.Members[0], targetChat.Members[1]
		uidARead := reads[uidA]
		uidBRead := reads[uidB]
		updateUids := []int64{}
		if !uidARead {
			updateUids = append(updateUids, uidA)
		}
		if !uidBRead {
			updateUids = append(updateUids, uidB)
		}

		err = s.chatInboxBiz.BatchDecrUnreadCount(ctx, updateUids, chatId)
		if err != nil {
			xlog.Msgf("batch decr unread count failed").
				Extras("chat_id", chatId, "msg_id", msgId, "uids", updateUids).Err(err).Errorx(ctx)
			// 更新失败不影响
		}
	}
	// TODO 此处人数太多的群聊可以先不处理 可能成本较大

	return nil
}

// 更新members收信箱
func (s *UserChatSrv) updateUserInboxes(ctx context.Context, chat *userchat.Chat,
	members []int64, msg *userchat.Msg) error {

	err := s.chatInboxBiz.BatchUpdateInboxLastMsgId(ctx, chat.Id, members, msg.Id)
	if err != nil {
		return xerror.Wrapf(err, "chat inbox biz batch update last_msg_id failed").WithCtx(ctx)
	}

	return nil
}

func (s *UserChatSrv) sendP2PMsg(ctx context.Context,
	sender int64, chat *userchat.Chat, msgReq *SendMsgReq, members []int64) (*userchat.Msg, error) {

	posKey := fmt.Sprintf("msger.userchat.chatmsg.pos:%s", chat.Id)
	res, err := dep.Idgen().GetId(ctx, posKey, 50)
	if err != nil {
		return nil, xerror.Wrapf(err, "idgen get id failed").WithCtx(ctx)
	}

	var (
		msgPos = int64(res)
		newMsg *userchat.Msg
	)

	err = infra.DaoTransact(ctx, func(ctx context.Context) error {
		// create msg
		newMsg, err = s.createMsg(ctx, sender, chat.Id, msgReq)
		if err != nil {
			return xerror.Wrapf(err, "create msg failed").WithCtx(ctx)
		}

		// bind msg to chat
		err = s.msgBiz.BindMsgToChat(ctx, newMsg.Id, chat.Id, msgPos)
		if err != nil {
			return xerror.Wrapf(err, "msg biz bind msg to chat failed").WithCtx(ctx)
		}

		// update chat's last msg
		err = s.chatBiz.UpdateChatLastMsg(ctx, chat.Id, newMsg.Id)
		if err != nil {
			return xerror.Wrapf(err, "chat biz update chat for msg failed").WithCtx(ctx)
		}

		return nil
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "user chat srv tx send p2p msg failed").WithCtx(ctx)
	}

	return newMsg, nil
}

// 发送群聊消息
func (s *UserChatSrv) sendGroupMsg(ctx context.Context,
	sender int64, chat *userchat.Chat, msgReq *SendMsgReq, members []int64) (*userchat.Msg, error) {
	panic("implement me") // TODO implement me
}

// sender创建一条消息
func (s *UserChatSrv) createMsg(ctx context.Context,
	sender int64, chatId uuid.UUID, msgReq *SendMsgReq) (*userchat.Msg, error) {

	newMsg, err := s.msgBiz.CreateMsg(ctx, sender, &userchat.CreateMsgReq{
		Type:    msgReq.Type,
		Content: msgReq.Content,
		Cid:     msgReq.Cid,
		Ext:     nil, // TODO implement me when supporting msg ext
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "msg biz create msg failed").WithCtx(ctx)
	}

	return newMsg, nil
}

func (s *UserChatSrv) isAllowedToSendMsg(ctx context.Context, sender int64, chat *userchat.Chat) error {
	canSend := false
	switch {
	case chat.IsP2PChat():
		if chat.IsUserInChat(sender) {
			canSend = true
		}
	case chat.IsGroupChat():
		// TODO
	default:
		return global.ErrUnsupportedChatType
	}

	if !canSend {
		return global.ErrUserNotInChat
	}

	return nil
}

func (s *UserChatSrv) isAllowedToRecallMsg(ctx context.Context,
	operator int64, chat *userchat.Chat, msg *userchat.Msg) error {

	// check msg time
	msgSendAt := msg.Id.Time()
	now := time.Now()
	if now.Sub(msgSendAt) > model.MaxRecallTime {
		return global.ErrRecallTimeReached
	}

	if !chat.IsUserInChat(operator) {
		return global.ErrCantRecallMsg
	}

	msgSender := msg.Sender
	if chat.IsP2PChat() {
		if msgSender != operator {
			return global.ErrCantRecallMsg
		}
	} else if chat.IsGroupChat() {
		// TODO
	} else {
		return global.ErrCantRecallMsg
	}

	return nil
}

// 检查uid的chatId收件箱是否已读某条信息
//
// 即检查uid是否已读chatId中的msgId
func (s *UserChatSrv) CheckUserInboxMsgRead(ctx context.Context, uid int64, chatId, msgId uuid.UUID) (bool, error) {
	logExtras := []any{"chat_id", chatId, "msg_id", msgId}
	chatInbox, err := s.chatInboxBiz.Get(ctx, uid, chatId)
	if err != nil {
		return false, xerror.Wrapf(err, "chat inbox biz get failed").
			WithExtras(logExtras...).WithCtx(ctx)
	}

	var (
		lastMsgId     = chatInbox.LastMsgId
		lastReadMsgId = chatInbox.LastReadMsgId
	)

	targetMsgIds := []uuid.UUID{lastMsgId, lastReadMsgId, msgId}
	msgPos, err := s.msgBiz.BatchGetMsgPos(ctx, chatId, targetMsgIds)
	if err != nil {
		return false, xerror.Wrapf(err, "msg biz batch get msg pos failed").
			WithExtras(logExtras...).WithCtx(ctx)
	}

	var (
		startPos, ok1  = msgPos[lastReadMsgId]
		endPos, ok2    = msgPos[lastMsgId]
		targetPos, ok3 = msgPos[msgId]
	)
	if !ok1 || !ok2 || !ok3 {
		err = global.ErrArgs.Msgf("msg pos missing: %v,%v,%v", ok1, ok2, ok3)
		return false, xerror.Wrap(err).
			WithExtras(logExtras...).
			WithCtx(ctx)
	}

	// if startPos < targetPos <= endPos, then msgId is considered unread by uid
	return !(startPos < targetPos && targetPos <= endPos), nil
}

// 检查uids是否已读chatId中msgId
func (s *UserChatSrv) BatchCheckUsersInBoxMsgRead(ctx context.Context, chatId, msgId uuid.UUID,
	uids []int64) (map[int64]bool, error) {

	logExtras := []any{"chat_id", chatId, "msg_id", msgId}

	// 获取每个uid的inbox的lastMsgId和lastReadMsgId
	inboxes, err := s.chatInboxBiz.BatchGet(ctx, chatId, uids)
	if err != nil {
		return nil, xerror.Wrapf(err, "chat inbox biz batch get failed").
			WithExtras(logExtras...).
			WithCtx(ctx)
	}

	inboxMsgIdsMap := make(map[uuid.UUID]struct{}, len(inboxes))
	for _, inbox := range inboxes {
		inboxMsgIdsMap[inbox.LastMsgId] = struct{}{}
		inboxMsgIdsMap[inbox.LastReadMsgId] = struct{}{}
	}

	// find all pos corrensponding to inboxMsgIds
	inboxMsgIds := xmap.Keys(inboxMsgIdsMap)
	inboxMsgIds = append(inboxMsgIds, msgId)
	inboxMsgPos, err := s.msgBiz.BatchGetMsgPos(ctx, chatId, inboxMsgIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "msg biz batch get msg pos failed").WithExtras(logExtras...).WithCtx(ctx)
	}

	readsResult := make(map[int64]bool, len(uids))
	for uid, inbox := range inboxes {
		lastMsgPos, ok1 := inboxMsgPos[inbox.LastMsgId]
		lastReadMsgPos, ok2 := inboxMsgPos[inbox.LastReadMsgId]
		targetMsgPos, ok3 := inboxMsgPos[msgId]
		if !ok1 || !ok2 || !ok3 {
			continue
		}

		// if startPos < targetPos <= endPos, then msgId is considered unread by uid
		unread := lastReadMsgPos < targetMsgPos && targetMsgPos <= lastMsgPos
		readsResult[uid] = !unread // 是否已读
	}

	return readsResult, nil
}
