package userchat

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
)

// 获取会话成员
func (s *UserChatSrv) GetChatMembers(ctx context.Context, chatId uuid.UUID) ([]int64, error) {
	chat, err := s.chatBiz.GetChat(ctx, chatId)
	if err != nil {
		return nil, xerror.Wrapf(err, "chat biz get chat failed").
			WithExtras("chat_id", chatId).WithCtx(ctx)
	}

	if chat.IsP2PChat() {
		return s.chatMemberBiz.GetP2PChatUsers(ctx, chatId)
	} else if chat.IsGroupChat() {
		// TODO
	}

	return nil, global.ErrChatNotExist
}

// 批量获取会话成员
func (s *UserChatSrv) BatchGetChatMembers(ctx context.Context,
	chatIds []uuid.UUID) (map[uuid.UUID][]int64, error) {
	return s.chatMemberBiz.BatchGetChatUsers(ctx, chatIds)
}
