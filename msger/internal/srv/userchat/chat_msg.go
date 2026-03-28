package userchat

import (
	"context"
	"math"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

// 获取消息列表
func (s *UserChatSrv) ListChatMsgs(ctx context.Context,
	chatId uuid.UUID, uid, pos int64, count int32,
	order model.Order,
) ([]*ChatMsg, error) {
	if order.Unspecified() {
		order = model.OrderDesc
	}

	if order.Desc() {
		// pos降序
		if pos <= 0 {
			pos = math.MaxInt64
		}
	} else if order.Asc() {
		// pos升序
		if pos <= 0 {
			pos = 0
		}
	}

	logAttrs := []any{"chat_id", chatId}

	// _, err := s.chatBiz.GetChat(ctx, chatId)
	// if err != nil {
	// 	return nil, xerror.Wrapf(err, "chat biz get chat failed").WithExtras(logAttrs...).WithCtx(ctx)
	// }

	uidInChat, err := s.chatMemberBiz.IsUserInChat(ctx, chatId, uid)
	if err != nil {
		return nil, xerror.Wrapf(err, "chat member biz check user in chat failed").WithExtras(logAttrs...).WithCtx(ctx)
	}
	if !uidInChat {
		return nil, xerror.Wrap(global.ErrUserNotInChat)
	}

	chatPos, err := s.msgBiz.ListChatPos(ctx, chatId, pos, count, order)
	if err != nil {
		return nil, xerror.Wrapf(err, "msg biz list chat pos failed").WithExtras(logAttrs...).WithCtx(ctx)
	}

	msgIds := make([]uuid.UUID, 0, len(chatPos))
	for _, cp := range chatPos {
		msgIds = append(msgIds, cp.MsgId)
	}

	msgs, err := s.msgBiz.BatchGetMsg(ctx, msgIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "msg biz batch get msg failed").WithExtras(logAttrs...).WithCtx(ctx)
	}

	// msgs需要和chatPos保持一样的顺序
	chatMsgs := make([]*ChatMsg, 0, len(msgs))
	for _, cp := range chatPos {
		chatMsgs = append(chatMsgs, &ChatMsg{
			Msg:    msgs[cp.MsgId],
			ChatId: chatId,
			Pos:    cp.Pos,
		})
	}

	return chatMsgs, nil
}
