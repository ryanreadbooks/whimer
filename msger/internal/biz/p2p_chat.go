package biz

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/msger/internal/infra"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dep"
)

// 用户单对单会话领域
type P2PChatBiz interface {
	// 两个用户开启会话
	InitiateChat(ctx context.Context, userA, userB int64) (int64, error)

	// 获取两个用户的会话
	GetChatByUsers(ctx context.Context, userA, userB int64) error
}

const (
	chatIdGenKey = "msger:p2p:chat:id"
)

type p2pChatBiz struct {
}

func NewP2PChatBiz() P2PChatBiz {
	return nil
}

// 两个用户开启会话
func (b *p2pChatBiz) InitiateChat(ctx context.Context, userA, userB int64) (int64, error) {
	// TODO 检查两个user的合法性

	seqNo, err := dep.Idgen().GetId(ctx, chatIdGenKey, 20000)
	if err != nil {
		return 0, xerror.Wrapf(err, "p2p biz failed to gen chatid").WithCtx(ctx)
	}

	chatId := int64(seqNo)

	err = infra.Dao().P2PChatDao.InitChat(ctx, chatId, userA, userB)
	if err != nil {
		return 0, xerror.Wrapf(err, "p2p biz failed to init chat").
			WithExtras("userA", userA, "userB", userB).
			WithCtx(ctx)
	}

	return chatId, nil
}
