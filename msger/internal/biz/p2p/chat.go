package p2p

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xmap"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	gm "github.com/ryanreadbooks/whimer/msger/internal/global/model"
	"github.com/ryanreadbooks/whimer/msger/internal/infra"
	p2pdao "github.com/ryanreadbooks/whimer/msger/internal/infra/dao/p2p"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dep"
)

type ListMsgReq struct {
	UserId int64
	ChatId int64
	Seq    int64
	Cnt    int32
	Unread bool // 是否仅列出未读
}

type ListChatReq struct {
	UserId     int64
	LastMsgSeq int64
	Count      int32
	Unread     bool // 是否仅列出未读的会话
}

const (
	chatIdGenKey  = "msger:p2p:chatid"
	chatIdGenStep = 20000
	msgIdGenKey   = "msger:p2p:msgid:%d:%d"
	msgIdGenStep  = 1000
)

type ChatBiz struct {
}

// 用户单对单会话领域
func NewP2PChatBiz() ChatBiz {
	return ChatBiz{}
}

// 两个用户开启会话, userA发起请求
func (b *ChatBiz) InitChat(ctx context.Context, userA, userB int64) (int64, error) {
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
func (b *ChatBiz) GetChatIdByUsers(ctx context.Context, userA, userB int64) (int64, error) {
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
func (b *ChatBiz) CreateMsg(ctx context.Context, req *CreateMsgReq) (*ChatMsg, error) {
	// 检查会话是否存在
	dualChats, err := infra.Dao().P2PChatDao.GetByChatId(ctx, req.ChatId)
	if err != nil {
		if errors.Is(err, xsql.ErrNoRecord) {
			return nil, xerror.Wrap(global.ErrP2PChatNotExist)
		}

		return nil, xerror.Wrapf(err, "p2p biz failed to get chat").WithExtra("req", req).WithCtx(ctx)
	}

	if len(dualChats) == 0 {
		return nil, xerror.Wrap(global.ErrP2PChatNotExist)
	}

	// 检查是否允许发送
	// 检查用户是否在会话中
	chatUid1, chatUid2 := dualChats[0].UserId, dualChats[0].PeerId
	if (req.Sender == chatUid1 && req.Receiver != chatUid2) ||
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
	seq := time.Now().UnixMicro()

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
		err = infra.Dao().P2PChatDao.UpdateMsg(ctx, msgId, msgPo.Seq, msgId, req.ChatId, req.Sender, false)
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

func (b *ChatBiz) getChatPO(ctx context.Context, userId, chatId int64) (*p2pdao.ChatPO, error) {
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

func (b *ChatBiz) GetMsg(ctx context.Context, chatId, msgId int64) (*ChatMsg, error) {
	r, err := infra.Dao().P2PMsgDao.GetByChatIdMsgId(ctx, chatId, msgId)
	if err != nil {
		return nil, xerror.Wrapf(err, "msg dao failed tod get msg").
			WithExtras("chat_id", chatId, "msg_id", msgId).WithCtx(ctx)
	}

	return MakeChatMsgFromPO(r, 0), nil
}

// 拉取userId在chatId中的会话信息(包含自己发送的和对方发送的)
func (b *ChatBiz) ListMsg(ctx context.Context, req *ListMsgReq) ([]*ChatMsg, error) {
	userChat, err := b.getChatPO(ctx, req.UserId, req.ChatId)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	boxMsgIds, err := infra.Dao().P2PInboxDao.ListMsg(ctx,
		req.UserId,
		req.ChatId,
		req.Seq,
		req.Cnt,
		req.Unread)
	if err != nil {
		return nil, xerror.Wrapf(err, "inbox dao failed to list inbox msg").WithCtx(ctx)
	}

	// 查消息
	return b.getChatMsgByIds(ctx, userChat, req.UserId, boxMsgIds)
}

// 查消息
func (b *ChatBiz) getChatMsgByIds(ctx context.Context, chat *p2pdao.ChatPO,
	userId int64, msgIds []int64) ([]*ChatMsg, error) {

	msgPos, err := infra.Dao().P2PMsgDao.GetByChatIdMsgIds(ctx, chat.ChatId, msgIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "msg dao failed to get by msg ids").WithCtx(ctx)
	}

	chatMsgs := make([]*ChatMsg, 0, len(msgPos))
	for _, msgPo := range msgPos {
		var recv int64 = userId
		if msgPo.SenderId == userId {
			recv = chat.PeerId
		}
		chatMsgs = append(chatMsgs, MakeChatMsgFromPO(msgPo, recv))
	}

	return chatMsgs, nil
}

// 获取用户会话的未读数
func (b *ChatBiz) GetUnreadCount(ctx context.Context, userId, chatId int64) (int64, error) {
	chat, err := b.getChatPO(ctx, userId, chatId)
	if err != nil {
		return 0, xerror.Wrapf(err, "p2p get unread count failed")
	}

	return chat.UnreadCount, nil
}

// 消除用户会话的未读数
func (b *ChatBiz) ClearUnreadCount(ctx context.Context, userId, chatId int64) error {
	_, err := b.getChatPO(ctx, userId, chatId)
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
func (b *ChatBiz) RevokeMessage(ctx context.Context, userId, chatId, msgId int64) error {
	// uid撤回chatId中msgId消息
	logExtras := make([]any, 0, 4)
	logExtras = append(logExtras, "chat_id", chatId, "msg_id", msgId, "user_id", userId)

	_, err := b.getChatPO(ctx, userId, chatId)
	if err != nil {
		return xerror.Wrapf(err, "p2p revoke message failed").WithExtras(logExtras...).WithCtx(ctx)
	}

	msgPo, err := infra.Dao().P2PMsgDao.GetByChatIdMsgId(ctx, chatId, msgId)
	if err != nil {
		return xerror.Wrapf(err, "msg dao failed to get msg").WithExtras(logExtras...).WithCtx(ctx)
	}
	if msgPo.SenderId != userId {
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

// 列出用户会话列表
func (b *ChatBiz) ListChat(ctx context.Context, req *ListChatReq) ([]*Chat, error) {
	lgExts := make([]any, 0, 4)
	lgExts = append(lgExts, "last_msg_seq", req.LastMsgSeq, "count", req.Count, "user_id", req.UserId)
	chatPos, err := infra.Dao().P2PChatDao.PageListByUserId(ctx,
		req.UserId,
		req.LastMsgSeq,
		int(req.Count),
		req.Unread)
	if err != nil {
		return nil, xerror.Wrapf(err, "chat dao failed to page list").WithExtras(lgExts...).WithCtx(ctx)
	}

	chats := make([]*Chat, 0, len(chatPos))
	for _, c := range chatPos {
		chats = append(chats, MakeChatFromPO(c))
	}

	if err := b.batchAssignLastMsg(ctx, chats); err != nil {
		return nil, xerror.Wrapf(err, "assign last msg failed").WithCtx(ctx)
	}

	return chats, nil
}

func (b *ChatBiz) batchAssignLastMsg(ctx context.Context, chats []*Chat) error {
	msgs, err := b.BatchGetLastMsg(ctx, chats)
	if err != nil {
		return xerror.Wrapf(err, "batch get last msg failed")
	}

	for _, chat := range chats {
		lastMsg := msgs[chat.ChatId]
		chat.LastMsg = lastMsg
	}

	return nil
}

func (b *ChatBiz) GetChat(ctx context.Context, userId, chatId int64) (*Chat, error) {
	chatPo, err := infra.Dao().P2PChatDao.GetByChatIdUserId(ctx, chatId, userId)
	if err != nil {
		return nil, xerror.Wrapf(err, "chat dao failed to get chat").
			WithExtras("chat_id", chatId, "user_id", userId).WithCtx(ctx)
	}

	chat := MakeChatFromPO(chatPo)
	if err := b.batchAssignLastMsg(ctx, []*Chat{chat}); err != nil {
		return nil, xerror.Wrapf(err, "assign last msg failed").WithCtx(ctx)
	}

	return chat, nil
}

func (b *ChatBiz) BatchGetChat(ctx context.Context, userId int64, chatIds []int64) (map[int64]*Chat, error) {
	chats, err := infra.Dao().P2PChatDao.BatchGetByChatIdsUserId(ctx, chatIds, userId)
	if err != nil {
		return nil, xerror.Wrapf(err, "chat dao failed to batch get chat").
			WithExtras("chat_ids", chatIds, "user_id", userId).WithCtx(ctx)
	}

	results := make(map[int64]*Chat, len(chats))
	for _, c := range chats {
		results[c.ChatId] = MakeChatFromPO(c)
	}

	resultChats := xmap.Values(results)
	if err := b.batchAssignLastMsg(ctx, resultChats); err != nil {
		return nil, xerror.Wrapf(err, "assign last msg failed").WithCtx(ctx)
	}
	return results, nil
}

// 批量获取会话的最近一条消息
func (b *ChatBiz) BatchGetLastMsg(ctx context.Context, chats []*Chat) (map[int64]*ChatMsg, error) {
	if len(chats) == 0 {
		return map[int64]*ChatMsg{}, nil
	}

	chatIds := make([]int64, 0, len(chats))
	msgIds := make([]int64, 0, len(chats))
	chatMap := make(map[int64]*Chat, len(chats))
	for _, c := range chats {
		chatIds = append(chatIds, c.ChatId)
		msgIds = append(msgIds, c.LastMsgId)
		chatMap[c.ChatId] = c
	}

	items, err := infra.Dao().P2PMsgDao.BatchGetByChatIdMsgId(ctx, chatIds, msgIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "msg dao failed to batch get").WithCtx(ctx)
	}

	// organize
	// chat_id -> []msg
	preResults := make(map[int64][]*ChatMsg)
	for _, m := range items {
		chat, ok := chatMap[m.ChatId]
		var recv int64
		if ok {
			if m.SenderId == chat.UserId {
				recv = chat.PeerId
			} else {
				recv = chat.UserId
			}
		}

		preResults[m.ChatId] = append(preResults[m.ChatId], MakeChatMsgFromPO(m, recv))
	}

	results := make(map[int64]*ChatMsg, len(preResults))
	// filter
	for chatId, pr := range preResults {
		sort.Slice(pr, func(i, j int) bool { return pr[i].Seq > pr[j].Seq })
		// get the first as result
		if len(pr) > 0 {
			results[chatId] = pr[0]
		}
	}

	return results, nil
}
