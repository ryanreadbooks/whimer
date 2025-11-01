package system

import (
	"context"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/infra"
	systemdao "github.com/ryanreadbooks/whimer/msger/internal/infra/dao/system"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

type ChatBiz struct{}

// 系统消息领域
func NewSystemChatBiz() ChatBiz {
	return ChatBiz{}
}

// 初始化用户的特定类型系统会话
func (b *ChatBiz) InitChat(ctx context.Context, userId int64, chatType model.SystemChatType) (uuid.UUID, error) {
	// 检查是否已存在该类型会话
	chat, err := infra.Dao().SystemChatDao.GetByUidAndType(ctx, userId, chatType)
	if err != nil {
		if !errors.Is(err, xsql.ErrNoRecord) {
			return uuid.UUID{}, xerror.Wrapf(err, "system chat dao failed to get chat").
				WithExtras("user_id", userId, "chat_type", chatType).
				WithCtx(ctx)
		}
		// 不存在则创建
		chatId := uuid.NewUUID()
		now := time.Now().UnixMicro()
		emptyUUID := uuid.EmptyUUID()
		err = infra.Dao().SystemChatDao.Create(ctx, &systemdao.ChatPO{
			Id:            chatId,
			Type:          chatType,
			Uid:           userId,
			Mtime:         now,
			LastMsgId:     emptyUUID,
			LastReadMsgId: emptyUUID,
			UnreadCount:   0,
		})
		if err != nil {
			return uuid.UUID{}, xerror.Wrapf(err, "system chat dao failed to create").
				WithExtras("user_id", userId, "chat_type", chatType).
				WithCtx(ctx)
		}
		return chatId, nil
	}

	return chat.Id, nil
}

// 发送系统消息
func (b *ChatBiz) CreateMsg(ctx context.Context, req *CreateSystemMsgReq) (*SystemMsg, error) {
	// 初始化系统会话
	chatId, err := b.InitChat(ctx, req.RecvUid, req.ChatType)
	if err != nil {
		return nil, xerror.Wrapf(err, "system chat biz failed to init chat").
			WithExtra("req", req).WithCtx(ctx)
	}

	msgId := uuid.NewUUID()
	now := time.Now().UnixMicro()

	msgPo := &systemdao.MsgPO{
		Id:           msgId,
		SystemChatId: chatId,
		Uid:          req.TriggerUid,
		RecvUid:      req.RecvUid,
		Status:       model.SystemMsgStatusNormal,
		MsgType:      req.MsgType,
		Content:      req.Content,
		Mtime:        now,
	}

	err = infra.DaoTransact(ctx, func(tctx context.Context) error {
		// 创建消息
		err := infra.Dao().SystemMsgDao.Create(tctx, msgPo)
		if err != nil {
			return xerror.Wrapf(err, "system msg dao failed to create")
		}

		// 更新会话的最后消息和未读数
		err = infra.Dao().SystemChatDao.UpdateLastMsg(tctx, chatId, msgId, true)
		if err != nil {
			return xerror.Wrapf(err, "system chat dao failed to update last msg")
		}

		return nil
	})

	if err != nil {
		return nil, xerror.Wrapf(err, "system biz failed to create msg").
			WithExtra("req", req).WithCtx(ctx)
	}

	resMsg := &SystemMsg{
		Id:           msgId,
		SystemChatId: chatId,
		TriggerUid:   req.TriggerUid,
		RecvUid:      req.RecvUid,
		Status:       model.SystemMsgStatusNormal,
		MsgType:      req.MsgType,
		Content:      req.Content,
		Mtime:        now,
	}

	return resMsg, nil
}

// 批量发送系统消息
func (b *ChatBiz) BatchCreateMsg(ctx context.Context, reqs []*CreateSystemMsgReq) ([]*SystemMsg, error) {
	msgs := make([]*SystemMsg, 0, len(reqs))
	msgPos := make([]*systemdao.MsgPO, 0, len(reqs))
	chatIds := make(map[int64]uuid.UUID)

	now := time.Now().UnixMicro()

	// 准备所有消息数据并确保会话存在
	for _, req := range reqs {
		chatId, ok := chatIds[req.RecvUid]
		if !ok {
			var err error
			chatId, err = b.InitChat(ctx, req.RecvUid, req.ChatType)
			if err != nil {
				return nil, xerror.Wrapf(err, "system chat biz failed to init chat").
					WithExtra("req", req).WithCtx(ctx)
			}
			chatIds[req.RecvUid] = chatId
		}

		msgId := uuid.NewUUID()
		msgPo := &systemdao.MsgPO{
			Id:           msgId,
			SystemChatId: chatId,
			Uid:          req.TriggerUid,
			RecvUid:      req.RecvUid,
			Status:       model.SystemMsgStatusNormal,
			MsgType:      req.MsgType,
			Content:      req.Content,
			Mtime:        now,
		}
		msgPos = append(msgPos, msgPo)

		msgs = append(msgs, &SystemMsg{
			Id:           msgId,
			SystemChatId: chatId,
			TriggerUid:   0,
			RecvUid:      req.RecvUid,
			Status:       model.SystemMsgStatusNormal,
			MsgType:      req.MsgType,
			Content:      req.Content,
			Mtime:        now,
		})
	}

	err := infra.Dao().DB().Transact(ctx, func(tctx context.Context) error {
		// 批量创建消息
		err := infra.Dao().SystemMsgDao.BatchCreate(tctx, msgPos)
		if err != nil {
			return xerror.Wrapf(err, "system msg dao failed to batch create")
		}

		// 更新每个会话的最后消息和未读数
		for recvUid, chatId := range chatIds {
			// 找到该用户的最新消息
			var latestMsgId uuid.UUID
			for _, msg := range msgPos {
				if msg.RecvUid == recvUid {
					latestMsgId = msg.Id
					break
				}
			}

			if !latestMsgId.EqualsTo(uuid.EmptyUUID()) {
				err = infra.Dao().SystemChatDao.UpdateLastMsg(tctx, chatId, latestMsgId, true)
				if err != nil {
					return xerror.Wrapf(err, "system chat dao failed to update last msg")
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, xerror.Wrapf(err, "system biz failed to batch create msg").WithCtx(ctx)
	}

	return msgs, nil
}

type ListMsgResp struct {
	Msgs    []*SystemMsg
	ChatId  uuid.UUID
	HasMore bool
}

func (b *ChatBiz) GetMsg(ctx context.Context, msgId uuid.UUID) (*SystemMsg, error) {
	m, err := infra.Dao().SystemMsgDao.GetById(ctx, msgId)
	if err != nil {
		return nil, xerror.Wrapf(err, "system msg dao get by id failed").WithExtra("msg_id", msgId).WithCtx(ctx)
	}

	return MakeSystemMsgFromPO(m), nil
}

func (b *ChatBiz) GetMsgChatId(ctx context.Context, msgId uuid.UUID) (uuid.UUID, error) {
	zeroUUID := uuid.EmptyUUID()
	chatId, err := infra.Dao().SystemMsgDao.GetChatIdById(ctx, msgId)
	if err != nil {
		if errors.Is(err, xsql.ErrNoRecord) {
			return zeroUUID, nil
		}

		return zeroUUID, xerror.Wrapf(err, "system msg dao get by id failed").
			WithExtra("msg_id", msgId).WithCtx(ctx)
	}
	return chatId, nil
}

// 分页获取用户的系统消息
func (b *ChatBiz) ListMsg(ctx context.Context, req *ListMsgReq) (*ListMsgResp, error) {
	// 获取会话
	resp := &ListMsgResp{Msgs: []*SystemMsg{}}
	chat, err := infra.Dao().SystemChatDao.GetByUidAndType(ctx, req.RecvUid, req.ChatType)
	if err != nil {
		if errors.Is(err, xsql.ErrNoRecord) {
			// 会话不存在，返回空列表
			return resp, nil
		}
		return nil, xerror.Wrapf(err, "system chat dao failed to get chat").
			WithExtras("user_id", req.RecvUid, "chat_type", req.ChatType).
			WithCtx(ctx)
	}

	cursor, err := uuid.ParseString(req.Cursor)
	if err != nil {
		cursor = uuid.MaxUUID()
	}

	// 查询消息
	msgPos, err := infra.Dao().SystemMsgDao.ListByChatId(ctx, chat.Id, cursor, req.Count+1)
	if err != nil {
		return nil, xerror.Wrapf(err, "system msg dao failed to list msg").WithCtx(ctx)
	}
	if len(msgPos) == 0 {
		return resp, nil
	}

	hasMore := false
	if len(msgPos) == int(req.Count)+1 {
		hasMore = true
		msgPos = msgPos[:len(msgPos)-1]
	}

	msgs := MakeSystemMsgsFromPOs(msgPos)
	resp.HasMore = hasMore
	resp.Msgs = msgs
	resp.ChatId = chat.Id

	return resp, nil
}

// 撤回系统消息
func (b *ChatBiz) RevokeMsg(ctx context.Context, msgId uuid.UUID) error {
	msgPo, err := infra.Dao().SystemMsgDao.GetById(ctx, msgId)
	if err != nil {
		// 消息不存在
		if errors.Is(err, xsql.ErrNoRecord) {
			return xerror.Wrap(global.ErrMsgNotExist).WithExtra("msg_id", msgId)
		}
		return xerror.Wrapf(err, "system msg dao failed to get msg").
			WithExtra("msg_id", msgId).WithCtx(ctx)
	}

	// 检查是否已撤回
	if msgPo.Status == model.SystemMsgStatusRevoked {
		return xerror.Wrapf(global.ErrMsgAlreadyRevoked,
			"the message has been revoked").WithExtra("msg_id", msgId)
	}

	// 检查是否超时
	if msgPo.Mtime+model.MaxRevokeTime.Microseconds() < time.Now().UnixMicro() {
		return xerror.Wrapf(global.ErrMsgRevokedTimeReached,
			"exceeded the maximum revocation time").WithExtra("msg_id", msgId)
	}

	err = infra.Dao().SystemMsgDao.UpdateStatus(ctx, msgId, model.SystemMsgStatusRevoked)
	if err != nil {
		return xerror.Wrapf(err, "system msg dao failed to update status").
			WithExtra("msg_id", msgId).WithCtx(ctx)
	}

	return nil
}

// 删除uid的chatId中的msgId系统消息
func (b *ChatBiz) DeleteMsg(ctx context.Context, chatId, msgId uuid.UUID, recvUid int64) error {
	var (
		logExtras = []any{"chat_id", chatId, "msg_id", msgId, "recv_uid", recvUid}
	)

	err := infra.DaoTransact(ctx, func(ctx context.Context) error {
		chat, err := infra.Dao().SystemChatDao.GetByIdForUpdate(ctx, chatId)
		if err != nil {
			if errors.Is(err, xsql.ErrNoRecord) {
				return nil
			}

			return xerror.Wrapf(err, "deleting msg get system chat by id failed")
		}

		// 先检查消息是否存在且属于该用户
		msgPo, err := infra.Dao().SystemMsgDao.GetByIdForUpdate(ctx, msgId)
		if err != nil {
			// 消息不存在
			if errors.Is(err, xsql.ErrNoRecord) {
				return nil
			}

			return xerror.Wrapf(err, "system msg dao failed to get msg")
		}

		// 检查是否是该用户的消息
		if msgPo.RecvUid != recvUid {
			return xerror.Wrapf(global.ErrCantRevokeMsg, "not the owner of the message")
		}

		err = infra.Dao().SystemMsgDao.DeleteByChatIdMsgIdRecvUid(ctx, chatId, msgId, recvUid)
		if err != nil {
			return xerror.Wrapf(err, "system msg dao failed to delete msg")
		}

		// 对应chat的处理
		var (
			curChatLastMsgId     = chat.LastMsgId
			curChatLastReadMsgId = chat.LastReadMsgId
			curChatUnread        = chat.UnreadCount

			newLastMsgId     = curChatLastMsgId
			newLastReadMsgId = curChatLastReadMsgId
			newUnreadCount   = curChatUnread
		)

		if curChatLastMsgId.EqualsTo(msgId) {
			// find last msg
			newLastMsg, err := infra.Dao().SystemMsgDao.GetLastMsg(ctx, chatId, recvUid)
			if err != nil {
				return xerror.Wrapf(err, "system msg dao get last failed")
			}

			newLastMsgId = newLastMsg.Id
		}

		if curChatLastReadMsgId.EqualsTo(msgId) {
			// find last status=read msg
			newLastReadMsg, err := infra.Dao().SystemMsgDao.GetLastReadMsg(ctx, chatId, recvUid)
			if err != nil {
				return xerror.Wrapf(err, "system msg dao get last read msg failed")
			}

			newLastReadMsgId = newLastReadMsg.Id
		}

		if curChatUnread != 0 && msgPo.Status.Unread() {
			newUnreadCount -= 1
			newUnreadCount = max(0, newUnreadCount)
		}

		if newLastMsgId != curChatLastMsgId ||
			newLastReadMsgId != curChatLastReadMsgId ||
			newUnreadCount != curChatUnread {
			err = infra.Dao().SystemChatDao.UpdateMsgs(ctx, chatId, newLastMsgId, newLastReadMsgId, newUnreadCount)
			if err != nil {
				return xerror.Wrapf(err, "system chat dao update msgs failed")
			}
		}

		return nil
	})
	if err != nil {
		return xerror.Wrapf(err, "delete msg transaction failed").WithExtras(logExtras...).WithCtx(ctx)
	}

	return err
}

// 获取uid的所有系统会话
func (b *ChatBiz) ListUserChats(ctx context.Context, uid int64) ([]*SystemChat, error) {
	chatPos, err := infra.Dao().SystemChatDao.ListByUid(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "system chat dao failed to list chat").
			WithExtra("uid", uid).WithCtx(ctx)
	}

	chats := make([]*SystemChat, 0, len(chatPos))
	for _, c := range chatPos {
		chats = append(chats, MakeSystemChatFromPO(c))
	}

	// 批量获取最后消息
	if err := b.batchAssignLastMsg(ctx, chats); err != nil {
		return nil, xerror.Wrapf(err, "batch assign last msg failed").WithCtx(ctx)
	}

	return chats, nil
}

// 获取用户的特定类型系统会话
func (b *ChatBiz) GetChat(ctx context.Context, uid int64, chatType model.SystemChatType) (*SystemChat, error) {
	chatPo, err := infra.Dao().SystemChatDao.GetByUidAndType(ctx, uid, chatType)
	if err != nil {
		// 会话不存在
		if errors.Is(err, xsql.ErrNoRecord) {
			return nil, xerror.Wrapf(err, "system chat not exist").
				WithExtras("uid", uid, "chat_type", chatType)
		}
		return nil, xerror.Wrapf(err, "system chat dao failed to get chat").
			WithExtras("uid", uid, "chat_type", chatType).WithCtx(ctx)
	}

	chat := MakeSystemChatFromPO(chatPo)
	if err := b.batchAssignLastMsg(ctx, []*SystemChat{chat}); err != nil {
		return nil, xerror.Wrapf(err, "assign last msg failed").WithCtx(ctx)
	}

	return chat, nil
}

// 获取用户会话的未读数
func (b *ChatBiz) GetUnreadCount(ctx context.Context, uid int64, chatType model.SystemChatType) (int64, error) {
	chat, err := infra.Dao().SystemChatDao.GetByUidAndType(ctx, uid, chatType)
	if err != nil {
		if errors.Is(err, xsql.ErrNoRecord) {
			return 0, nil
		}
		return 0, xerror.Wrapf(err, "system chat dao failed to get chat").
			WithExtras("uid", uid, "chat_type", chatType).WithCtx(ctx)
	}

	return chat.UnreadCount, nil
}

// 清除特定会话的未读数
func (b *ChatBiz) ClearChatUnread(ctx context.Context, uid int64, chatId uuid.UUID) error {
	err := infra.DaoTransact(ctx, func(ctx context.Context) error {
		chat, err := infra.Dao().SystemChatDao.GetByIdForUpdate(ctx, chatId)
		if err != nil {
			if errors.Is(err, xsql.ErrNoRecord) {
				return xerror.Wrap(global.ErrSysChatNotExist)
			}

			return xerror.Wrapf(err, "system chat dao failed to get by id").
				WithExtra("chat_id", chatId).WithCtx(ctx)
		}

		if chat.Uid != uid {
			return xerror.Wrap(global.ErrSysChatNotYours)
		}

		if err := b.clearUnread(ctx, uid, chat); err != nil {
			return xerror.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return xerror.Wrapf(err, "clear chat unread transaction failed")
	}

	return nil
}

// 清空uid会话的未读数 同时需要把会话的最后已读消息ID设置为最后消息ID 并且更新msg的status
func (b *ChatBiz) ClearUidUnread(ctx context.Context, uid int64, chatType model.SystemChatType) error {
	err := infra.DaoTransact(ctx, func(ctx context.Context) error {
		chat, err := infra.Dao().SystemChatDao.GetByUidAndTypeForUpdate(ctx, uid, chatType)
		if err != nil {
			if errors.Is(err, xsql.ErrNoRecord) {
				return nil
			}
			return xerror.Wrapf(err, "system chat dao failed to get chat").
				WithExtras("uid", uid, "chat_type", chatType).WithCtx(ctx)
		}

		if err := b.clearUnread(ctx, uid, chat); err != nil {
			return xerror.Wrap(err)
		}

		return nil
	})
	if err != nil {
		return xerror.Wrapf(err, "clear uid unread transaction failed")
	}

	return nil
}

func (b *ChatBiz) clearUnread(ctx context.Context, uid int64, chat *systemdao.ChatPO) error {
	// 获取会话最新一条消息
	lastMsg, err := infra.Dao().SystemMsgDao.GetLastMsg(ctx, chat.Id, uid)
	if err != nil {
		return xerror.Wrapf(err, "system msg dao failed to get last msg").WithCtx(ctx)
	}

	// 更新chat
	err = infra.Dao().SystemChatDao.ClearUnread(ctx, chat.Id, lastMsg.Id)
	if err != nil {
		return xerror.Wrapf(err, "system chat dao failed to clear unread").WithCtx(ctx)
	}

	// 批量更新msg状态
	err = infra.Dao().SystemMsgDao.UpdateStatusToTarget(ctx,
		chat.Id, uid, model.SystemMsgStatusRead, model.SystemMsgStatusNormal)
	if err != nil {
		return xerror.Wrapf(err, "system msg dao failed to update status").WithCtx(ctx)
	}

	return nil
}

// 批量获取会话的最后消息
func (b *ChatBiz) batchAssignLastMsg(ctx context.Context, chats []*SystemChat) error {
	if len(chats) == 0 {
		return nil
	}

	// 收集需要查询的消息ID
	msgIds := make([]uuid.UUID, 0, len(chats))
	chatMap := make(map[uuid.UUID]*SystemChat, len(chats))

	for _, chat := range chats {
		if !chat.LastMsgId.EqualsTo(uuid.EmptyUUID()) {
			msgIds = append(msgIds, chat.LastMsgId)
			chatMap[chat.Id] = chat
		}
	}

	if len(msgIds) == 0 {
		return nil
	}

	// 批量获取消息
	msgPos, err := infra.Dao().SystemMsgDao.BatchGetByIds(ctx, msgIds)
	if err != nil {
		return xerror.Wrapf(err, "system msg dao failed to batch get").WithCtx(ctx)
	}

	// 构建消息映射
	msgMap := make(map[uuid.UUID]*systemdao.MsgPO, len(msgPos))
	for _, msgPo := range msgPos {
		msgMap[msgPo.Id] = msgPo
	}

	// 为每个会话设置最后消息
	for _, chat := range chats {
		if !chat.LastMsgId.EqualsTo(uuid.EmptyUUID()) {
			if msgPo, ok := msgMap[chat.LastMsgId]; ok {
				chat.LastMsg = MakeSystemMsgFromPO(msgPo)
			}
		}
	}

	return nil
}

func (b *ChatBiz) GetChatUnreadCount(ctx context.Context, uid int64, chatId uuid.UUID) (*ChatUnread, error) {
	// get chat
	chat, err := infra.Dao().SystemChatDao.GetById(ctx, chatId)
	if err != nil {
		if errors.Is(err, xsql.ErrNoRecord) {
			return nil, xerror.Wrap(global.ErrSysChatNotExist)
		}
		return nil, xerror.Wrapf(err, "system chat dao failed to get by id").
			WithExtra("chat_id", chatId).WithCtx(ctx)
	}

	if chat.Uid != uid {
		return nil, xerror.Wrap(global.ErrSysChatNotYours)
	}

	return ChatUnreadFromPo(chat), nil
}

func (b *ChatBiz) GetUserAllChatUnreadCount(ctx context.Context, uid int64) ([]*ChatUnread, error) {
	chats, err := infra.Dao().SystemChatDao.ListByUid(ctx, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "system chat dao list by uid failed").
			WithExtra("uid", uid).WithCtx(ctx)
	}

	cu := make([]*ChatUnread, 0, len(chats))
	for _, c := range chats {
		cu = append(cu, ChatUnreadFromPo(c))
	}

	return cu, nil
}
