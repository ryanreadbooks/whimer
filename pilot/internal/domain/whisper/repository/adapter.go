package repository

import (
	"context"

	"github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/vo"
)

type SendMsgParams struct {
	Sender  int64
	ChatId  string
	Cid     string
	Type    vo.MsgType
	Content *vo.MsgContent
}

type ListRecentChatsResult struct {
	Chats      []*entity.RecentChat
	NextCursor string
	HasNext    bool
}

type UserChatAdapter interface {
	CreateP2PChat(ctx context.Context, uid, target int64) (chatId string, err error)
	SendMsgToChat(ctx context.Context, params *SendMsgParams) (msgId string, err error)
	GetChatMembers(ctx context.Context, chatId string) ([]int64, error)
	BatchGetChatMembers(ctx context.Context, chatIds []string) (map[string][]int64, error)
	ListRecentChats(ctx context.Context, uid int64, cursor string, count int32) (*ListRecentChatsResult, error)
	ListChatMsgs(ctx context.Context, chatId string, uid int64, pos int64, count int32) ([]*entity.Msg, error)
	RecallMsg(ctx context.Context, uid int64, chatId, msgId string) error
	ClearChatUnread(ctx context.Context, uid int64, chatId string) error
}
