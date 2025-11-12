package userchat

import (
	"context"
	"slices"

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

func (b *ChatMemberBiz) BatchGetChatUsers(ctx context.Context, chatIds []uuid.UUID) (map[uuid.UUID][]int64, error) {
	p2pChats := make([]uuid.UUID, 0, len(chatIds))
	groupChats := make([]uuid.UUID, 0, len(chatIds))

	chats, err := infra.Dao().ChatDao.BatchGetById(ctx, chatIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "chat dao batch get failed").WithCtx(ctx)
	}

	for _, chat := range chats {
		switch chat.Type {
		case model.P2PChat:
			p2pChats = append(p2pChats, chat.Id)
		case model.GroupChat:
			groupChats = append(groupChats, chat.Id)
		}
	}

	p2pResults, err := infra.Dao().ChatMemberP2PDao.BatchGetByChatId(ctx, p2pChats)
	if err != nil {
		return nil, xerror.Wrapf(err, "chat member p2p dao batch get failed").WithCtx(ctx)
	}

	// group results

	members := make(map[uuid.UUID][]int64, len(p2pChats)+len(groupChats))
	for _, p2p := range p2pResults {
		members[p2p.ChatId] = append(members[p2p.ChatId], []int64{p2p.UidA, p2p.UidB}...)
	}

	return members, nil
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
		if len(chat.Members) != 2 {
			return xerror.Wrap(global.ErrInternal.Msg("p2p chat members is not of length 2"))
		}
	case model.GroupChat:
		// TODO
	}

	return nil
}

func (b *ChatMemberBiz) IsUserInChat(ctx context.Context, chatId uuid.UUID, uid int64) (bool, error) {
	chatMembers, err := b.BatchGetChatUsers(ctx, []uuid.UUID{chatId})
	if err != nil {
		return false, xerror.Wrapf(err, "batch get chat users failed").WithCtx(ctx)
	}

	members, ok := chatMembers[chatId]
	if !ok {
		return false, xerror.Wrap(global.ErrChatNotExist)
	}

	if !slices.Contains(members, uid) {
		// uid not in chat
		return false, nil
	}

	return true, nil
}
