package userchat

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	bizuserchat "github.com/ryanreadbooks/whimer/msger/internal/biz/userchat"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
	"golang.org/x/sync/errgroup"
)

// 列出最近会话列表
func (s *UserChatSrv) ListRecentChats(ctx context.Context, uid int64,
	cursor string, count int32) ([]*RecentChat, *model.PageListResult[string], error) {

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

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(recovery.DoV2(func() error {
		var err error
		chats, err = s.chatBiz.BatchGetChat(ctx, chatIds)
		if err != nil {
			return xerror.Wrapf(err, "chat biz batch get chat failed").WithExtras(logAttrs...).WithCtx(ctx)
		}
		return nil
	}))

	eg.Go(recovery.DoV2(func() error {
		var err error
		msgs, err = s.msgBiz.BatchGetMsg(ctx, lastMsgIds)
		if err != nil {
			return xerror.Wrapf(err, "msg biz batch get msg failed").WithExtras(logAttrs...).WithCtx(ctx)
		}
		return nil
	}))
	err = eg.Wait()
	if err != nil {
		return nil, pageResult, err
	}

	// organize
	recentChats := make([]*RecentChat, 0, len(inboxes))
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

		recentChats = append(recentChats, &RecentChat{
			Uid:           inbox.Uid,
			ChatId:        inbox.ChatId,
			ChatType:      inboxChat.Type,
			ChatName:      inboxChat.Name,
			ChatStatus:    inboxChat.Status,
			ChatCreator:   inboxChat.Creator,
			LastMsg:       inboxLastMsg,
			LastReadMsgId: inbox.LastReadMsgId,
			LastReadTime:  inbox.LastReadTime,
			UnreadCount:   inbox.UnreadCount,
			Ctime:         inbox.Ctime,
			Mtime:         inbox.Mtime,
			IsPinned:      inbox.IsPinned,
		})
	}

	return recentChats, pageResult, nil
}
