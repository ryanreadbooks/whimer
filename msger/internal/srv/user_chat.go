package srv

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/msger/internal/biz"
	"github.com/ryanreadbooks/whimer/msger/internal/biz/userchat"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/infra"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dep"
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

	lockKey := fmt.Sprintf("msger:userchat:lock:initp2pchat%d:%d", userA, userB)
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
// 1. 创建消息
// 2. 消息绑定会话
// 3. 更新会话最后一条消息
// 4. 更新所有user的收件箱, 所有user的收件箱的插入新信件(更新lastMsgId)
func (s *UserChatSrv) SendMsg(ctx context.Context,
	sender int64, chatId uuid.UUID, msgReq *SendMsgReq) error {

	targetChat, err := s.chatBiz.GetChat(ctx, chatId)
	if err != nil {
		return xerror.Wrapf(err, "chat biz get chat failed").WithCtx(ctx)
	}

	// check if user can send to chat
	chatMembers, err := s.isAllowedToSendMsg(ctx, sender, targetChat)
	if err != nil {
		return xerror.Wrapf(err, "sender unable to send").
			WithExtras("sender", sender).WithCtx(ctx)
	}

	// TODO check chat status and settings
	if !targetChat.IsStatusNormal() {
		return xerror.Wrap(global.ErrChatNotNormal)
	}

	// 异步准备收件箱
	eg, gctx := errgroup.WithContext(ctx)
	eg.Go(recovery.DoV2(func() error {
		err := s.chatInboxBiz.BatchPrepareInboxes(gctx, chatId, chatMembers)
		if err != nil {
			return xerror.Wrapf(err, "async chat inbox biz batch prepare failed").WithCtx(gctx)
		}

		return nil
	}))

	var newMsg *userchat.Msg
	switch {
	case targetChat.IsP2PChat():
		newMsg, err = s.sendP2PMsg(ctx, sender, targetChat, msgReq, chatMembers)
	case targetChat.IsGroupChat():
		newMsg, err = s.sendGroupMsg(ctx, sender, targetChat, msgReq, chatMembers)
	default:
		return xerror.Wrap(global.ErrUnsupportedChatType)
	}

	if err != nil {
		return xerror.Wrapf(err, "user chat send msg failed").WithCtx(ctx)
	}

	gerr := eg.Wait()
	if gerr != nil {
		xlog.Msg("send msg async work failed").Err(err).Errorx(ctx)
		// do prepare again
		err := s.chatInboxBiz.BatchPrepareInboxes(gctx, chatId, chatMembers)
		if err != nil {
			return xerror.Wrapf(err, "chat inbox biz batch prepare failed").WithCtx(ctx)
		}
	}

	// 写入信箱
	err = s.updateUserInboxes(ctx, targetChat, chatMembers, newMsg)
	if err != nil {
		return xerror.Wrapf(err, "update user inboxes failed").WithCtx(ctx)
	}

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

func (s *UserChatSrv) isAllowedToSendMsg(ctx context.Context, sender int64, chat *userchat.Chat) ([]int64, error) {

	canSend := false
	var members []int64
	switch {
	case chat.IsP2PChat():
		var err error
		members, err = s.chatMemberBiz.GetP2PChatUsers(ctx, chat.Id)
		if err != nil {
			return nil, xerror.Wrapf(err, "get p2p chat users err").WithCtx(ctx)
		}
		if slices.Contains(members, sender) {
			canSend = true
		}
	case chat.IsGroupChat():
		// TODO
	default:
		return nil, xerror.Wrap(global.ErrUnsupportedChatType)
	}

	if !canSend {
		return nil, xerror.Wrap(global.ErrUserNotInChat)
	}

	return members, nil
}
