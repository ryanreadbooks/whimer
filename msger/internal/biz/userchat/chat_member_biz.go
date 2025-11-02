package userchat

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/infra"
	"github.com/ryanreadbooks/whimer/msger/internal/infra/dao/chat"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

type ChatMemberBiz struct {
}

func NewChatMemberBiz() ChatMemberBiz {
	return ChatMemberBiz{}
}

// 单聊会话创建成员
func (b *ChatMemberBiz) InsertP2PMembers(ctx context.Context, chatId uuid.UUID, uidA, uidB int64) error {
	err := infra.Dao().ChatMemberP2PDao.Create(ctx, &chat.ChatMemberP2PPO{
		ChatId: chatId,
		UidA:   uidA,
		UidB:   uidB,
		Ctime:  getNormalTime(),
		Mtime:  getNormalTime(),
	})
	if err != nil {
		return xerror.Wrapf(err, "chat member p2p dao create failed").WithCtx(ctx)
	}

	return nil
}

func (b *ChatMemberBiz) GetP2PChatId(ctx context.Context, uidA, uidB int64) (uuid.UUID, error) {
	members, err := infra.Dao().ChatMemberP2PDao.GetByUids(ctx, uidA, uidB)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return uuid.EmptyUUID(), global.ErrChatNotExist
		}

		return uuid.EmptyUUID(), xerror.Wrapf(err, "chat member p2p dao get by uids failed").
			WithExtras("uid_a", uidA, "uid_b", uidB).
			WithCtx(ctx)
	}

	return members.ChatId, nil
}

// 获取单聊两个用户
func (b *ChatMemberBiz) GetP2PChatUsers(ctx context.Context, chatId uuid.UUID) ([]int64, error) {
	members, err := infra.Dao().ChatMemberP2PDao.GetByChatId(ctx, chatId)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return nil, global.ErrChatNotExist
		}

		return nil, xerror.Wrapf(err, "chat member p2p dao get by chatid failed").
			WithExtras("chat_id", chatId).
			WithCtx(ctx)
	}

	return []int64{members.UidA, members.UidB}, nil
}

func (b *ChatMemberBiz) AttachChatMembers(ctx context.Context, chat *Chat) error {
	if chat == nil {
		return nil
	}

	switch chat.Type {
	case model.P2PChat:
		members, err := b.GetP2PChatUsers(ctx, chat.Id)
		if err != nil {
			return xerror.Wrapf(err, "get p2p chat users err").WithCtx(ctx)
		}
		chat.Members = members
	case model.GroupChat:
		// TODO
	}

	return nil
}
