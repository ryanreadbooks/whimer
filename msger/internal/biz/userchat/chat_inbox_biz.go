package userchat

import (
	"context"
	"sync"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/infra"
	chatdao "github.com/ryanreadbooks/whimer/msger/internal/infra/dao/chat"
)

type ChatInboxBiz struct {
}

func NewChatInboxBiz() ChatInboxBiz {
	return ChatInboxBiz{}
}

func (b *ChatInboxBiz) Get(ctx context.Context, uid int64, chatId uuid.UUID) (*ChatInbox, error) {
	po, err := infra.Dao().ChatInboxDao.GetByUidChatId(ctx, uid, chatId)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return nil, global.ErrChatInboxNotExist
		}
		return nil, xerror.Wrapf(err, "chat inbox dao get by uid chatid failed").
			WithExtras("req_uid", uid, "chat_id", chatId).WithCtx(ctx)
	}

	return makeChatInboxFromPO(po), nil
}

func (b *ChatInboxBiz) BatchGet(ctx context.Context,
	chatId uuid.UUID, uids []int64) (map[int64]*ChatInbox, error) {
	inboxesPo, err := infra.Dao().ChatInboxDao.BatchGetChatIdUids(ctx, chatId, uids)
	if err != nil {
		return nil, xerror.Wrapf(err, "chat inbox dao batch get failed").
			WithExtras("req_uid", uids, "chat_id", chatId).WithCtx(ctx)
	}

	result := make(map[int64]*ChatInbox, len(inboxesPo))
	for uid, inbox := range inboxesPo {
		result[uid] = makeChatInboxFromPO(inbox)
	}

	return result, nil
}

// 创建用户uid的chatId信箱
func (b *ChatInboxBiz) PrepareInbox(ctx context.Context, uid int64, chatId uuid.UUID) error {
	now := getAccurateTime()
	err := infra.Dao().ChatInboxDao.Create(ctx, &chatdao.ChatInboxPO{
		Uid:    uid,
		ChatId: chatId,
		Ctime:  now,
		Mtime:  now,
	})
	if err != nil {
		return xerror.Wrapf(err, "chat inbox dao create failed").WithCtx(ctx)
	}

	return nil
}

// 初始化用户uids在chatId的信箱 存在则忽略
func (b *ChatInboxBiz) BatchPrepareInboxes(ctx context.Context, chatId uuid.UUID, uids []int64) error {
	now := getAccurateTime()
	bss := make([]*chatdao.ChatInboxPO, 0, len(uids))
	for _, uid := range uids {
		bss = append(bss, &chatdao.ChatInboxPO{
			Uid:    uid,
			ChatId: chatId,
			Ctime:  now,
			Mtime:  now,
		})
	}

	err := infra.Dao().ChatInboxDao.BatchCreate(ctx, bss)
	if err != nil {
		return xerror.Wrapf(err, "chat inbox dao batch create failed").WithCtx(ctx)
	}

	return nil
}

func (b *ChatInboxBiz) CheckInboxExist(ctx context.Context, uid int64, chatId uuid.UUID) (bool, error) {
	_, err := infra.Dao().ChatInboxDao.GetByUidChatId(ctx, uid, chatId)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return false, nil
		}

		return false, xerror.Wrapf(err, "chat inbox dao get by uid chatid failed").
			WithExtras("req_uid", uid, "chat_id", chatId).WithCtx(ctx)
	}

	return true, nil
}

// 批量更新uid信箱最后一条msgId
func (b *ChatInboxBiz) BatchUpdateInboxLastMsgId(ctx context.Context,
	chatId uuid.UUID, uids []int64, msgId uuid.UUID) error {

	var wg sync.WaitGroup
	err := xslice.BatchAsyncExec(&wg, uids, 100, func(start, end int) error {
		targetUids := uids[start:end]
		now := getAccurateTime()
		err := infra.Dao().ChatInboxDao.BatchUpdateLastMsgId(ctx, chatId, targetUids, msgId, now)
		if err != nil {
			return xerror.Wrapf(err, "chat inbox dao batch update last_msg_id failed").WithCtx(ctx)
		}

		return nil
	})

	if err != nil {
		// 仅打日志
		xlog.Msg("chat inbox dao batch update last_msg_id has error").Err(err).Errorx(ctx)
	}

	return nil
}

// 将uid信箱的最后已读消息设置为最后一条消息
func (b *ChatInboxBiz) SetLastReadMsgIdToLatest(ctx context.Context, chatId uuid.UUID, uid int64) error {
	err := infra.Dao().ChatInboxDao.SetLastReadMsgId(ctx, uid, chatId, getAccurateTime())
	if err != nil {
		return xerror.Wrapf(err, "chat inbox dao set last_read_msg_id failed").
			WithExtras("chat_id", chatId).
			WithCtx(ctx)
	}

	return nil
}
