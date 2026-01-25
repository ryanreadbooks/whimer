package repository

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"
)

type SystemNotifyAdapter interface {
	NotifyLikesMsg(ctx context.Context, msg *vo.SystemMessage) (string, error)
	NotifyReplyMsg(ctx context.Context, msg *vo.SystemMessage) (string, error)
	NotifyMentionMsg(ctx context.Context, msg []*vo.SystemMessage) (map[int64][]string, error)
}
