package userchat

import (
	"context"
	"sync"

	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	bizuserchat "github.com/ryanreadbooks/whimer/msger/internal/biz/userchat"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
	"golang.org/x/sync/errgroup"
)

// 列出最近会话列表
func (s *UserChatSrv) ListRecentChats(ctx context.Context, uid int64, cursor string, count int32) (
	[]*RecentChat, *model.PageListResult[string], error) {

	logAttrs := []any{"uid", uid, "cursor", cursor}

	inboxes, pageResult, err := s.chatInboxBiz.ListByUid(ctx, uid, cursor, count)
	if err != nil {
		return nil, nil, xerror.Wrapf(err, "chat inbox biz list by uid failed").WithExtras(logAttrs...).WithCtx(ctx)
	}

	// 获取chat
	chatIds := make([]uuid.UUID, 0, len(inboxes))
	lastMsgIds := make([]uuid.UUID, 0, len(inboxes))
	for _, inbox := range inboxes {
		chatIds = append(chatIds, inbox.ChatId)
		// omit zero id
		if !inbox.LastMsgId.IsZero() {
			lastMsgIds = append(lastMsgIds, inbox.LastMsgId)
		}
	}

	var (
		chats map[uuid.UUID]*bizuserchat.Chat
		msgs  map[uuid.UUID]*bizuserchat.Msg
	)

	eg, ctx2 := errgroup.WithContext(ctx)
	eg.Go(recovery.DoV2(func() error {
		var err error
		chats, err = s.chatBiz.BatchGetChat(ctx2, chatIds)
		if err != nil {
			return xerror.Wrapf(err, "chat biz batch get chat failed").WithExtras(logAttrs...).WithCtx(ctx2)
		}
		return nil
	}))

	eg.Go(recovery.DoV2(func() error {
		var err error
		msgs, err = s.msgBiz.BatchGetMsg(ctx2, lastMsgIds)
		if err != nil {
			return xerror.Wrapf(err, "msg biz batch get msg failed").WithExtras(logAttrs...).WithCtx(ctx2)
		}
		return nil
	}))
	err = eg.Wait() // wait会cancel ctx 所以用一个新变量ctx2
	if err != nil {
		return nil, pageResult, err
	}

	// organize
	recentChats := make([]*RecentChat, 0, len(inboxes))
	chatMsgs := make([]*ChatMsg, 0, len(recentChats))

	for _, inbox := range inboxes {
		inboxChat, ok := chats[inbox.ChatId]
		if !ok {
			// abort
			return nil, pageResult,
				xerror.Wrapf(global.ErrListRecentChatNoChatId, "inbox %s no chat", inbox.ChatId).WithCtx(ctx)
		}
		inboxLastMsg, ok := msgs[inbox.LastMsgId]
		if !ok {
			inboxLastMsg = &bizuserchat.Msg{}
		}

		inboxLastChatMsg := makeChatMsgFromMsg(inboxLastMsg)
		chatMsgs = append(chatMsgs, inboxLastChatMsg)
		inboxLastChatMsg.ChatId = inbox.ChatId

		recentChats = append(recentChats, &RecentChat{
			Uid:           inbox.Uid,
			ChatId:        inbox.ChatId,
			ChatType:      inboxChat.Type,
			ChatName:      inboxChat.Name,
			ChatStatus:    inboxChat.Status,
			ChatCreator:   inboxChat.Creator,
			LastMsg:       inboxLastChatMsg,
			LastReadMsgId: inbox.LastReadMsgId,
			LastReadTime:  inbox.LastReadTime,
			UnreadCount:   inbox.UnreadCount,
			Ctime:         inbox.Ctime,
			Mtime:         inbox.Mtime,
			IsPinned:      inbox.IsPinned,
		})
	}

	// 绑定pos
	err = s.bindChatMsgPos(ctx, chatMsgs)
	if err != nil {
		return nil, pageResult, xerror.Wrapf(err, "bind chat msg pos failed").WithCtx(ctx)
	}

	return recentChats, pageResult, nil
}

// chatIdsMsgIds: chatId -> []msgIds
func (s *UserChatSrv) bindChatMsgPos(ctx context.Context, chatMsgs []*ChatMsg) error {
	chatIdsMsgIds := make(map[uuid.UUID][]uuid.UUID)
	for _, chatMsg := range chatMsgs {
		chatIdsMsgIds[chatMsg.ChatId] = append(chatIdsMsgIds[chatMsg.ChatId], chatMsg.Id)
	}

	var (
		mu                    sync.Mutex
		chatIdMsgIdPosMapping = make(map[uuid.UUID]map[uuid.UUID]int64)
	)

	eg, ctx := errgroup.WithContext(ctx)
	for chatId, msgIds := range chatIdsMsgIds {
		eg.Go(recovery.DoV2(func() error {
			msgPoses, err := s.msgBiz.BatchGetMsgPos(ctx, chatId, msgIds)
			if err != nil {
				return xerror.Wrapf(err, "msg biz batch get msg pos failed").
					WithExtras("chat_id", chatId, "msg_ids", msgIds).WithCtx(ctx)
			}

			mu.Lock()
			chatIdMsgIdPosMapping[chatId] = msgPoses
			mu.Unlock()

			return nil
		}))
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	for _, chatMsg := range chatMsgs {
		if posMapping, ok := chatIdMsgIdPosMapping[chatMsg.ChatId]; ok {
			if pos, ok := posMapping[chatMsg.Id]; ok {
				chatMsg.Pos = pos
			}
		}
	}

	return nil
}
