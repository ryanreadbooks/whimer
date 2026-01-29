package whisper

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	userchatv1 "github.com/ryanreadbooks/whimer/msger/api/userchat/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/whisper/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/whisper/convert"
)

var _ repository.UserChatAdapter = (*UserChatAdapterImpl)(nil)

type UserChatAdapterImpl struct {
	client userchatv1.UserChatServiceClient
}

func NewUserChatAdapterImpl(client userchatv1.UserChatServiceClient) *UserChatAdapterImpl {
	return &UserChatAdapterImpl{client: client}
}

func (a *UserChatAdapterImpl) CreateP2PChat(ctx context.Context, uid, target int64) (string, error) {
	resp, err := a.client.CreateP2PChat(ctx,
		&userchatv1.CreateP2PChatRequest{
			Uid:    uid,
			Target: target,
		})
	if err != nil {
		return "", xerror.Wrapf(err, "create p2p chat failed").WithCtx(ctx)
	}
	return resp.ChatId, nil
}

func (a *UserChatAdapterImpl) SendMsgToChat(ctx context.Context, params *repository.SendMsgParams) (string, error) {
	pbMsgReq := convert.SendMsgParamsToPb(params)

	resp, err := a.client.SendMsgToChat(ctx,
		&userchatv1.SendMsgToChatRequest{
			Sender: params.Sender,
			ChatId: params.ChatId,
			Msg:    pbMsgReq,
		})
	if err != nil {
		return "", xerror.Wrapf(err, "send msg to chat failed").WithCtx(ctx)
	}
	return resp.MsgId, nil
}

func (a *UserChatAdapterImpl) GetChatMembers(ctx context.Context, chatId string) ([]int64, error) {
	resp, err := a.client.GetChatMembers(ctx,
		&userchatv1.GetChatMembersRequest{
			ChatId: chatId,
		})
	if err != nil {
		return nil, xerror.Wrapf(err, "get chat members failed").WithCtx(ctx)
	}
	return resp.GetMembers(), nil
}

func (a *UserChatAdapterImpl) BatchGetChatMembers(ctx context.Context, chatIds []string) (map[string][]int64, error) {
	resp, err := a.client.BatchGetChatMembers(ctx,
		&userchatv1.BatchGetChatMembersRequest{
			ChatIds: chatIds,
		})
	if err != nil {
		return nil, xerror.Wrapf(err, "batch get chat members failed").WithCtx(ctx)
	}

	result := make(map[string][]int64, len(resp.GetMembersMap()))
	for chatId, members := range resp.GetMembersMap() {
		result[chatId] = members.GetInts()
	}
	return result, nil
}

func (a *UserChatAdapterImpl) ListRecentChats(
	ctx context.Context, uid int64, cursor string, count int32,
) (*repository.ListRecentChatsResult, error) {
	resp, err := a.client.ListRecentChats(ctx,
		&userchatv1.ListRecentChatsRequest{
			Uid:    uid,
			Cursor: cursor,
			Count:  count,
		})
	if err != nil {
		return nil, xerror.Wrapf(err, "list recent chats failed").WithCtx(ctx)
	}

	chats := make([]*entity.RecentChat, 0, len(resp.RecentChats))
	for _, pbChat := range resp.RecentChats {
		chats = append(chats, convert.RecentChatFromPb(pbChat))
	}

	return &repository.ListRecentChatsResult{
		Chats:      chats,
		NextCursor: resp.NextCursor,
		HasNext:    resp.HasNext,
	}, nil
}

func (a *UserChatAdapterImpl) ListChatMsgs(
	ctx context.Context, chatId string, uid int64, pos int64, count int32,
) ([]*entity.Msg, error) {
	resp, err := a.client.ListChatMsgs(ctx,
		&userchatv1.ListChatMsgsRequest{
			ChatId: chatId,
			Uid:    uid,
			Pos:    pos,
			Count:  count,
		})
	if err != nil {
		return nil, xerror.Wrapf(err, "list chat msgs failed").WithCtx(ctx)
	}

	msgs := make([]*entity.Msg, 0, len(resp.GetChatMsgs()))
	for _, pbMsg := range resp.GetChatMsgs() {
		msgs = append(msgs, convert.MsgFromPb(pbMsg))
	}
	return msgs, nil
}

func (a *UserChatAdapterImpl) RecallMsg(ctx context.Context, uid int64, chatId, msgId string) error {
	_, err := a.client.RecallMsg(ctx,
		&userchatv1.RecallMsgRequest{
			Uid:    uid,
			MsgId:  msgId,
			ChatId: chatId,
		})
	if err != nil {
		return xerror.Wrapf(err, "recall msg failed").WithCtx(ctx).WithExtras("msg_id", msgId, "chat_id", chatId)
	}
	return nil
}

func (a *UserChatAdapterImpl) ClearChatUnread(ctx context.Context, uid int64, chatId string) error {
	_, err := a.client.ClearChatUnread(ctx,
		&userchatv1.ClearChatUnreadRequest{
			ChatId: chatId,
			Uid:    uid,
		})
	if err != nil {
		return xerror.Wrapf(err, "clear chat unread failed").WithCtx(ctx).WithExtras("chat_id", chatId)
	}
	return nil
}
