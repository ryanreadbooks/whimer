package userchat

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/msger/internal/global"
	"github.com/ryanreadbooks/whimer/msger/internal/infra"
	chatdao "github.com/ryanreadbooks/whimer/msger/internal/infra/dao/chat"
	"github.com/ryanreadbooks/whimer/msger/internal/model"
)

type ChatBiz struct {
}

func NewChatBiz() ChatBiz {
	return ChatBiz{}
}

// 创建单聊会话
func (b *ChatBiz) CreateP2PChat(ctx context.Context) (uuid.UUID, error) {
	chatId := uuid.NewUUID()
	c := chatdao.ChatPO{
		Id:     chatId,
		Type:   model.P2PChat,
		Status: model.ChatStatusNormal,
		Mtime:  getAccurateTime(),
	}

	err := infra.Dao().ChatDao.Create(ctx, &c)
	if err != nil {
		return chatId, xerror.Wrapf(err, "chat dao create failed").WithCtx(ctx)
	}

	return chatId, nil
}

// 创建群聊会话
func (b *ChatBiz) CreateGroupChat(ctx context.Context, name string, creator int64) (uuid.UUID, error) {
	chatId := uuid.NewUUID()
	c := chatdao.ChatPO{
		Id:      chatId,
		Type:    model.GroupChat,
		Status:  model.ChatStatusNormal,
		Mtime:   getAccurateTime(),
		Name:    name,
		Creator: creator,
	}

	err := infra.Dao().ChatDao.Create(ctx, &c)
	if err != nil {
		return chatId, xerror.Wrapf(err, "chat dao create failed").WithCtx(ctx)
	}

	return chatId, nil
}

// 获取会话
func (b *ChatBiz) GetChat(ctx context.Context, chatId uuid.UUID) (*Chat, error) {
	po, err := infra.Dao().ChatDao.GetById(ctx, chatId)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return nil, global.ErrChatNotExist
		}
		return nil, xerror.Wrapf(err, "chat dao get failed").WithCtx(ctx)
	}

	return makeChatFromPO(po), nil
}

func (b *ChatBiz) BatchGetChat(ctx context.Context, chatIds []uuid.UUID) (map[uuid.UUID]*Chat, error) {
	if len(chatIds) == 0 {
		return make(map[uuid.UUID]*Chat), nil
	}

	chatPoes, err := infra.Dao().ChatDao.BatchGetById(ctx, chatIds)
	if err != nil {
		return nil, xerror.Wrapf(err, "chat dao batch get failed").WithCtx(ctx)
	}

	result := make(map[uuid.UUID]*Chat, len(chatPoes))
	for _, chat := range chatPoes {
		result[chat.Id] = makeChatFromPO(chat)
	}

	return result, nil
}

// 消息发送时更新会话
func (b *ChatBiz) UpdateChatLastMsg(ctx context.Context, chatId, msgId uuid.UUID) error {
	err := infra.Dao().ChatDao.UpdateLastMsgId(ctx, chatId, msgId, getAccurateTime())
	if err != nil {
		return xerror.Wrapf(err, "chat dao update last msg id failed").WithCtx(ctx)
	}

	return nil
}
