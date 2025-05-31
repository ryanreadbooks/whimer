package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/msger/internal/biz"
	bizp2p "github.com/ryanreadbooks/whimer/msger/internal/biz/p2p"
)

type P2PChatSrv struct {
	chatBiz bizp2p.ChatBiz
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

// 单聊消息发送
func (s *P2PChatSrv) SendMessage(ctx context.Context, req *bizp2p.CreateMsgReq) (
	*bizp2p.ChatMsg, error) {
	// TODO 数据推送下发
	return s.chatBiz.CreateMsg(ctx, req)
}

// 获取会话消息
func (s *P2PChatSrv) ListMessage(ctx context.Context, userId, chatId, seq int64, cnt int32) (
	[]*bizp2p.ChatMsg, int64, error) {
	msgs, err := s.chatBiz.ListMsg(ctx, userId, chatId, seq, cnt)
	if err != nil {
		return nil, 0, err
	}
	if len(msgs) == 0 {
		return nil, 0, nil
	}

	nextSeq := msgs[len(msgs)-1].Seq

	return msgs, nextSeq, nil
}

// 获取未读数
func (s *P2PChatSrv) GetUnread(ctx context.Context, userId, chatId int64) (int64, error) {
	return s.chatBiz.GetUnreadCount(ctx, userId, chatId)
}

// 清除未读数
func (s *P2PChatSrv) ClearUnread(ctx context.Context, userId, chatId int64) error {
	return s.chatBiz.ClearUnreadCount(ctx, userId, chatId)
}

// 撤回消息
func (s *P2PChatSrv) RevokeMessage(ctx context.Context, chatId, msgId int64) error {
	// TODO 数据推送下发
	return s.chatBiz.RevokeMessage(ctx, chatId, msgId)
}
