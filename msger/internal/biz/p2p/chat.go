package p2p

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
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

	// TODO 检查是否允许发送

	uid1 := min(req.Sender, req.Receiver)
	uid2 := max(req.Sender, req.Receiver)
	msgNo, err := dep.Idgen().GetId(ctx, fmt.Sprintf(msgIdGenKey, uid1, uid2), msgIdGenStep)
	if err != nil {
		return nil, xerror.Wrapf(err, "p2p biz failed to get msgid").WithExtra("req", req).WithCtx(ctx)
	}

	msgId := int64(msgNo)
	seq := time.Now().UnixNano()

	msgPo := &p2pdao.Message{
		MsgId:    msgId,
		SenderId: req.Sender,
		ChatId:   req.ChatId,
		MsgType:  int8(req.MsgType),
		Status:   int8(MsgStatusNormal),
		Seq:      seq,
		Utime:    seq,
		Content:  req.Content, // TODO content需要加密
	}

	err = infra.Dao().DB().Transact(ctx, func(ctx context.Context) error {
		// 创建消息
		err = infra.Dao().P2PMsgDao.Create(ctx, msgPo)
		if err != nil {
			return xerror.Wrapf(err, "p2p msg dao failed to create")
		}

		// 写入inbox
		senderInbox := p2pdao.InboxMsg{
			UserId: req.Sender,
			ChatId: msgPo.ChatId,
			MsgId:  msgId,
			Status: int8(InboxRead),
		}
		receiverInbox := p2pdao.InboxMsg{
			UserId: req.Receiver,
			ChatId: msgPo.ChatId,
			MsgId:  msgId,
			Status: int8(InboxUnread),
		}
		err := infra.Dao().P2PInboxDao.BatchCreate(ctx, []*p2pdao.InboxMsg{&senderInbox, &receiverInbox})
		if err != nil {
			return xerror.Wrapf(err, "p2p inbox dao failed to batch create")
		}

		// 更新receiver的未读数
		err = infra.Dao().P2PChatDao.UpdateLastMsg(ctx, msgId, msgPo.Seq, req.ChatId, req.Receiver, true)
		if err != nil {
			return xerror.Wrapf(err, "p2p chat dao failed to update last msg for sender")
		}

		// 更新sender的消息
		err = infra.Dao().P2PChatDao.UpdateLastMsg(ctx, msgId, msgPo.Seq, req.ChatId, req.Sender, false)
		if err != nil {
			return xerror.Wrapf(err, "p2p chat dao failed to update last msg for sender")
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
		Status:   MsgStatus(msgPo.Status),
		Content:  req.Content,
		Seq:      seq,
	}

	return resChatMsg, nil
}
