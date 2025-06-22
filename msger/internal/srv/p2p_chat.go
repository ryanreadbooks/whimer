package srv

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/msger/internal/biz"
	bizp2p "github.com/ryanreadbooks/whimer/msger/internal/biz/p2p"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dep"
	userv1 "github.com/ryanreadbooks/whimer/passport/api/user/v1"

	"golang.org/x/sync/errgroup"
)

type P2PChatSrv struct {
	chatBiz bizp2p.ChatBiz
}

func NewP2PChatSrv(biz biz.Biz) *P2PChatSrv {
	return &P2PChatSrv{
		chatBiz: biz.P2PBiz,
	}
}

func (s *P2PChatSrv) GetChat(ctx context.Context, userId, chatId int64) (*bizp2p.Chat, error) {
	return s.chatBiz.GetChat(ctx, userId, chatId)
}

func (s *P2PChatSrv) checkUsers(ctx context.Context, uid1, uid2 int64) error {
	eg, _ := errgroup.WithContext(ctx)
	for _, user := range []int64{uid1, uid2} {
		var userId = user
		eg.Go(func() error {
			resp, err := dep.Userer().HasUser(ctx, &userv1.HasUserRequest{Uid: userId})
			if err != nil {
				return xerror.Wrapf(err, "get user failed").WithCtx(ctx)
			}
			if resp.GetHas() {
				return nil
			}
			return global.ErrUserNotFound
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

// 两个用户创建会话
func (s *P2PChatSrv) CreateChat(ctx context.Context, initer, target int64) (int64, error) {
	if err := s.checkUsers(ctx, initer, target); err != nil {
		return 0, err
	}

	chatId, err := s.chatBiz.InitChat(ctx, initer, target)
	if err != nil {
		return 0, xerror.Wrapf(err, "p2p chat srv failed to init chat").WithCtx(ctx)
	}

	return chatId, nil
}

// 单聊消息发送
func (s *P2PChatSrv) SendMessage(ctx context.Context, req *bizp2p.CreateMsgReq) (
	*bizp2p.ChatMsg, error) {
	// 发送接收检查
	if err := s.checkUsers(ctx, req.Sender, req.Receiver); err != nil {
		return nil, err
	}

	msg, err := s.chatBiz.CreateMsg(ctx, req)
	if err != nil {
		return nil, xerror.Wrapf(err, "p2p chat srv failed to create msg").WithCtx(ctx)
	}

	s.notifyReceiver(ctx, req.Receiver)

	return msg, nil
}

// 获取会话消息
func (s *P2PChatSrv) ListMessage(ctx context.Context, userId, chatId, seq int64, cnt int32) (
	[]*bizp2p.ChatMsg, int64, error) {
	msgs, err := s.chatBiz.ListMsg(ctx, &bizp2p.ListMsgReq{
		UserId: userId,
		ChatId: chatId,
		Seq:    seq,
		Cnt:    cnt,
	})
	if err != nil {
		return nil, 0, err
	}
	var nextSeq int64 = -1
	if lc := len(msgs); lc != 0 && lc == int(cnt) {
		nextSeq = msgs[len(msgs)-1].Seq
	}

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
// userId撤回在chatId中的msgId消息
func (s *P2PChatSrv) RevokeMessage(ctx context.Context, userId, chatId, msgId int64) error {
	cht, err := s.chatBiz.GetChat(ctx, userId, chatId)
	if err != nil {
		return xerror.Wrapf(err, "p2p chat failed to get chat")
	}

	err = s.chatBiz.RevokeMsg(ctx, userId, chatId, msgId)
	if err != nil {
		return xerror.Wrapf(err, "p2p chat failed to revoke message")
	}

	s.notifyReceiver(ctx, cht.PeerId)

	return nil
}

// 列出会话列表
func (s *P2PChatSrv) ListChat(ctx context.Context, userId, seq int64, count int32) (
	[]*bizp2p.Chat, int64, error) {

	chats, err := s.chatBiz.ListChat(ctx, &bizp2p.ListChatReq{
		UserId:     userId,
		LastMsgSeq: seq,
		Count:      count,
	})
	if err != nil {
		return nil, 0, xerror.Wrapf(err, "p2p chat failed to list chat")
	}

	var nextSeq int64 = -1
	if lc := len(chats); lc != 0 && lc == int(count) {
		nextSeq = chats[len(chats)-1].LastMsgSeq
	}

	return chats, nextSeq, nil
}

type UnreadChat struct {
	*bizp2p.Chat
	*bizp2p.ChatMsg
}

// 列出未读会话 并且每个会话返回最新的未读消息
func (s *P2PChatSrv) ListUnreadChat(ctx context.Context, userId, seq int64, count int32) ([]*UnreadChat, int64, error) {
	chats, err := s.chatBiz.ListChat(ctx, &bizp2p.ListChatReq{
		UserId:     userId,
		LastMsgSeq: seq,
		Count:      count,
		Unread:     true,
	})
	if err != nil {
		return nil, 0, xerror.Wrapf(err, "p2p chat failed to list unread chat")
	}

	// chatid -> msg
	lastMsgs, err := s.chatBiz.BatchGetLastMsg(ctx, chats)
	if err != nil {
		return nil, 0, xerror.Wrapf(err, "p2p chat failed to batch get last msg")
	}

	var nextSeq int64 = -1
	if lc := len(chats); lc != 0 && lc == int(count) {
		nextSeq = chats[len(chats)-1].LastMsgSeq
	}

	results := make([]*UnreadChat, 0, len(chats))
	for _, c := range chats {
		msg := lastMsgs[c.ChatId]
		results = append(results, &UnreadChat{
			Chat:    c,
			ChatMsg: msg,
		})
	}

	return nil, nextSeq, nil
}

func (s *P2PChatSrv) DeleteChat(ctx context.Context, userId, chatId int64) error {
	return s.chatBiz.DeleteChat(ctx, userId, chatId)
}

func (s *P2PChatSrv) DeleteChatMessage(ctx context.Context, userId, chatId, msgId int64) error {
	return s.chatBiz.DeleteMsg(ctx, userId, chatId, msgId)
}
