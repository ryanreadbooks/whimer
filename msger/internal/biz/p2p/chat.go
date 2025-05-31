package p2p

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	gm "github.com/ryanreadbooks/whimer/msger/internal/global/model"
	"github.com/ryanreadbooks/whimer/msger/internal/infra"
	p2pdao "github.com/ryanreadbooks/whimer/msger/internal/infra/dao/p2p"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dep"
)

// 用户单对单会话领域
type ChatBiz interface {
	// 两个用户开启会话
	InitChat(ctx context.Context, userA, userB int64) (int64, error)
	// 获取两个用户的会话
	GetChatIdByUsers(ctx context.Context, userA, userB int64) (int64, error)
	// 发送消息
	CreateMsg(ctx context.Context, req *CreateMsgReq) (*ChatMsg, error)
	// 列出用户的会话消息
	ListMsg(ctx context.Context, userId, chatId, seq int64, cnt int32) ([]*ChatMsg, error)
	// 获取用户会话的未读数
	GetUnreadCount(ctx context.Context, userId, chatId int64) (int64, error)
	// 消除用户会话的未读数
	ClearUnreadCount(ctx context.Context, userId, chatId int64) error
	// 撤回会话消息
	RevokeMessage(ctx context.Context, chatId, msgId int64) error
}

const (
	chatIdGenKey  = "msger:p2p:chatid"
	chatIdGenStep = 20000
	msgIdGenKey   = "msger:p2p:msgid:%d:%d"
	msgIdGenStep  = 1000
)

type p2pChatBiz struct {
}

func NewP2PChatBiz() ChatBiz {
	return &p2pChatBiz{}
}

// 两个用户开启会话, userA发起请求
func (b *p2pChatBiz) InitChat(ctx context.Context, userA, userB int64) (int64, error) {
	// TODO 检查两个user的合法性

	seqNo, err := dep.Idgen().GetId(ctx, chatIdGenKey, chatIdGenStep)
	if err != nil {
		return 0, xerror.Wrapf(err, "p2p biz failed to gen chatid").WithCtx(ctx)
	}

	chatId := int64(seqNo)

	err = infra.Dao().P2PChatDao.InitChat(ctx, chatId, userA, userB)
	if err != nil {
		if !errors.Is(err, xsql.ErrDuplicate) {
			return 0, xerror.Wrapf(err, "p2p biz failed to init chat").
				WithExtras("userA", userA, "userB", userB).
				WithCtx(ctx)
		}

		// 会话已经创建了 直接查出来
		cid, err := b.GetChatIdByUsers(ctx, userA, userB)
		if err != nil {
			return 0, err
		}
		chatId = cid
	}

	return chatId, nil
}

// 获取两个用户的会话id
func (b *p2pChatBiz) GetChatIdByUsers(ctx context.Context, userA, userB int64) (int64, error) {
	chat, err := infra.Dao().P2PChatDao.GetByUsers(ctx, userA, userB)
	if err != nil {
		if errors.Is(err, xsql.ErrNoRecord) {
			return 0, xerror.Wrap(global.ErrP2PChatNotExist)
		}

		return 0, xerror.Wrapf(err, "p2p biz failed to get chat").
			WithExtras("userA", userA, "userB", userB).
			WithCtx(ctx)
	}

	return chat.ChatId, nil
}

// 发送消息
func (b *p2pChatBiz) CreateMsg(ctx context.Context, req *CreateMsgReq) (*ChatMsg, error) {
	// 检查会话是否存在
	dualChats, err := infra.Dao().P2PChatDao.GetByChatId(ctx, req.ChatId)
	if err != nil {
		if errors.Is(err, xsql.ErrNoRecord) {
			return nil, xerror.Wrap(global.ErrP2PChatNotExist)
		}

		return nil, xerror.Wrapf(err, "p2p biz failed to get chat").WithExtra("req", req).WithCtx(ctx)
	}

	if len(dualChats) != 2 {
		return nil, xerror.Wrap(global.ErrP2PChatNotExist)
	}

	// 检查是否允许发送
	// 检查用户是否在会话中
	chatUid1, chatUid2 := dualChats[0].UserId, dualChats[0].PeerId
	if (req.Sender == chatUid1 && req.Receiver == chatUid2) ||
		(req.Sender == chatUid2 && req.Receiver != chatUid1) {
		return nil, global.ErrUserNotInChat
	}

	// TODO 其它检查

	uid1 := min(chatUid1, chatUid2)
	uid2 := max(chatUid1, chatUid2)
	msgNo, err := dep.Idgen().GetId(ctx, fmt.Sprintf(msgIdGenKey, uid1, uid2), msgIdGenStep)
	if err != nil {
		return nil, xerror.Wrapf(err, "idgen failed to get msgid").WithExtra("req", req).WithCtx(ctx)
	}

	msgId := int64(msgNo)
	seq := time.Now().UnixNano()

	msgPo := &p2pdao.MessagePO{
		MsgId:    msgId,
		SenderId: req.Sender,
		ChatId:   req.ChatId,
		MsgType:  req.MsgType,
		Status:   gm.MsgStatusNormal,
		Seq:      seq,
		Utime:    seq,
		Content:  req.Content, // TODO content需要加密
	}

	err = infra.Dao().DB().Transact(ctx, func(ctx context.Context) error {
		// 创建消息
		err = infra.Dao().P2PMsgDao.Create(ctx, msgPo)
		if err != nil {
			return xerror.Wrapf(err, "msg dao failed to create")
		}

		// 写入inbox
		senderInbox := p2pdao.InboxMsgPO{
			UserId: req.Sender,
			ChatId: msgPo.ChatId,
			MsgId:  msgId,
			MsgSeq: seq,
			Status: gm.InboxRead,
		}
		receiverInbox := p2pdao.InboxMsgPO{
			UserId: req.Receiver,
			ChatId: msgPo.ChatId,
			MsgId:  msgId,
			MsgSeq: seq,
			Status: gm.InboxUnread,
		}
		err := infra.Dao().P2PInboxDao.BatchCreate(ctx, []*p2pdao.InboxMsgPO{&senderInbox, &receiverInbox})
		if err != nil {
			return xerror.Wrapf(err, "inbox dao failed to batch create")
		}

		// 更新receiver的未读数
		err = infra.Dao().P2PChatDao.UpdateLastMsg(ctx, msgId, msgPo.Seq, req.ChatId, req.Receiver, true)
		if err != nil {
			return xerror.Wrapf(err, "chat dao failed to update last msg for sender")
		}

		// 更新sender的消息
		err = infra.Dao().P2PChatDao.UpdateLastMsg(ctx, msgId, msgPo.Seq, req.ChatId, req.Sender, false)
		if err != nil {
			return xerror.Wrapf(err, "chat dao failed to update last msg for sender")
		}

		return nil
	})

	if err != nil {
		return nil, xerror.Wrapf(err, "p2p biz failed to create msg").WithExtra("req", req).WithCtx(ctx)
	}

	resChatMsg := &ChatMsg{
		MsgId:    msgId,
		Sender:   req.Sender,
		Receiver: req.Receiver,
		ChatId:   msgPo.ChatId,
		Type:     req.MsgType,
		Status:   msgPo.Status,
		Content:  req.Content,
		Seq:      seq,
	}

	return resChatMsg, nil
}

func (b *p2pChatBiz) getChat(ctx context.Context, userId, chatId int64) (*p2pdao.ChatPO, error) {
	c, err := infra.Dao().P2PChatDao.GetByChatIdUserId(ctx, chatId, userId)
	if err != nil {
		if errors.Is(err, xsql.ErrNoRecord) {
			return nil, global.ErrP2PChatNotExist
		}
		return nil, xerror.Wrapf(err, "chat dao failed to get chat").
			WithExtras("user_id", userId, "chat_id", chatId).WithCtx(ctx)
	}

	return c, nil
}

// 拉取userId在chatId中的会话信息(包含自己发送的和对方发送的)
func (b *p2pChatBiz) ListMsg(ctx context.Context,
	userId, chatId, seq int64, cnt int32) ([]*ChatMsg, error) {
	userChat, err := b.getChat(ctx, userId, chatId)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	boxMsgIds, err := infra.Dao().P2PInboxDao.ListMsg(ctx, userId, chatId, seq, cnt)
	if err != nil {
		return nil, xerror.Wrapf(err, "inbox dao failed to list inbox msg").WithCtx(ctx)
	}

	// 查消息
	msgPos, err := infra.Dao().P2PMsgDao.GetByMsgIds(ctx, chatId, boxMsgIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "msg dao failed to get by msg ids").WithCtx(ctx)
	}

	chatMsgs := make([]*ChatMsg, 0, len(msgPos))
	for _, msgPo := range msgPos {
		var recv int64 = userId
		if msgPo.SenderId == userId {
			recv = userChat.PeerId
		}
		chatMsgs = append(chatMsgs, MakeChatMsgFromPO(msgPo, recv))
	}

	return chatMsgs, nil
}

// 获取用户会话的未读数
func (b *p2pChatBiz) GetUnreadCount(ctx context.Context, userId, chatId int64) (int64, error) {
	chat, err := b.getChat(ctx, userId, chatId)
	if err != nil {
		return 0, xerror.Wrapf(err, "p2p get unread count failed")
	}

	return chat.UnreadCount, nil
}

// 消除用户会话的未读数
func (b *p2pChatBiz) ClearUnreadCount(ctx context.Context, userId, chatId int64) error {
	_, err := b.getChat(ctx, userId, chatId)
	if err != nil {
		return xerror.Wrapf(err, "p2p clear unread count failed")
	}

	err = infra.Dao().DB().Transact(ctx, func(tctx context.Context) error {
		// 清除chat中的未读数
		err := infra.Dao().P2PChatDao.ResetUnreadCount(tctx, chatId, userId)
		if err != nil {
			return xerror.Wrapf(err, "chat dao reset unread failed")
		}

		// 清除inbox中的未读
		err = infra.Dao().P2PInboxDao.UpdateStatusToRead(ctx, userId, chatId)
		if err != nil {
			return xerror.Wrapf(err, "inbox box failed to update status")
		}

		return nil
	})
	if err != nil {
		return xerror.Wrapf(err, "clear unread count dao transact failed").
			WithExtras("user_id", userId, "chat_id", chatId).WithCtx(ctx)
	}

	return nil
}

// 撤回消息
func (b *p2pChatBiz) RevokeMessage(ctx context.Context, chatId, msgId int64) error {
	uid := metadata.Uid(ctx)
	// uid撤回chatId中msgId消息
	logExtras := make([]any, 0, 4)
	logExtras = append(logExtras, "chat_id", chatId, "msg_id", msgId)

	_, err := b.getChat(ctx, uid, chatId)
	if err != nil {
		return xerror.Wrapf(err, "p2p revoke message failed").WithExtras(logExtras...).WithCtx(ctx)
	}

	msgPo, err := infra.Dao().P2PMsgDao.GetByMsgId(ctx, msgId)
	if err != nil {
		return xerror.Wrapf(err, "msg dao failed to get msg").WithExtras(logExtras...).WithCtx(ctx)
	}
	if msgPo.SenderId != uid {
		return global.ErrCantRevokeMsg
	}
	if msgPo.Status == gm.MsgStatusRevoked {
		return global.ErrMsgAlreadyRevoked
	}

	// 撤回 改消息表和对应的inbox
	err = infra.Dao().DB().Transact(ctx, func(tctx context.Context) error {
		// 改消息表
		err := infra.Dao().P2PMsgDao.RevokeMsg(ctx, chatId, msgId)
		if err != nil {
			return xerror.Wrapf(err, "msg dao revoke failed")
		}

		err = infra.Dao().P2PInboxDao.RevokeMsg(ctx, chatId, msgId)
		if err != nil {
			return xerror.Wrapf(err, "inbox dao revoke failed")
		}

		return nil
	})
	if err != nil {
		return xerror.Wrapf(err, "revoke msg dao transact failed").
			WithExtras(logExtras...).WithCtx(ctx)
	}

	return nil
}
