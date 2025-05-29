package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/msger/internal/biz"
	"github.com/ryanreadbooks/whimer/msger/internal/biz/p2p"
)

type P2PChatSrv struct {
	chatBiz p2p.ChatBiz
}

func NewP2PChatSrv(biz biz.Biz) *P2PChatSrv {
	return &P2PChatSrv{
		chatBiz: biz.P2PBiz,
	}
}

// 两个用户创建会话
func (s *P2PChatSrv) CreateChat(ctx context.Context, initer, target int64) (int64, error) {
	chatId, err := s.chatBiz.InitChat(ctx, initer, target)
	if err != nil {
		return 0, xerror.Wrapf(err, "p2p chat srv failed to init chat").WithCtx(ctx)
	}

	return chatId, nil
}
