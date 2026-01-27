package repository

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"
)

// SystemNotifyAdapter 系统通知适配器接口
type SystemNotifyAdapter interface {
	// 发送通知
	NotifyLikesMsg(ctx context.Context, msg *vo.SystemMessage) (string, error)
	NotifyReplyMsg(ctx context.Context, msg *vo.SystemMessage) (string, error)
	NotifyMentionMsg(ctx context.Context, msg []*vo.SystemMessage) (map[int64][]string, error)

	// 获取消息列表
	ListMentionMsg(ctx context.Context, uid int64, cursor string, count int32) (*vo.ListMsgResult, error)
	ListReplyMsg(ctx context.Context, uid int64, cursor string, count int32) (*vo.ListMsgResult, error)
	ListLikesMsg(ctx context.Context, uid int64, cursor string, count int32) (*vo.ListMsgResult, error)

	// 会话管理
	GetChatUnread(ctx context.Context, uid int64) (*entity.ChatsUnreadCount, error)
	ClearChatUnread(ctx context.Context, uid int64, chatId string) error
	DeleteMsg(ctx context.Context, uid int64, msgId string) error
}
