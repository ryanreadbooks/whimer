package systemnotify

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/adapter/systemnotify/convert"

	v1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
)

type SystemNotifyAdapterImpl struct {
	notifyCli v1.NotificationServiceClient
	chatCli   v1.ChatServiceClient
}

var _ repository.SystemNotifyAdapter = (*SystemNotifyAdapterImpl)(nil)

func NewSystemNotifyAdapterImpl(
	notifyCli v1.NotificationServiceClient,
	chatCli v1.ChatServiceClient,
) *SystemNotifyAdapterImpl {
	return &SystemNotifyAdapterImpl{
		notifyCli: notifyCli,
		chatCli:   chatCli,
	}
}

func (a *SystemNotifyAdapterImpl) NotifyLikesMsg(ctx context.Context, msg *vo.SystemMessage) (string, error) {
	resp, err := a.notifyCli.NotifyLikesMsg(ctx,
		&v1.NotifyLikesMsgRequest{
			Contents: []*v1.LikeMsgContent{{
				Uid:       msg.Uid,
				TargetUid: msg.TargetUid,
				Content:   msg.Content,
			}},
		})
	if err != nil {
		return "", xerror.Wrap(err)
	}

	msgIds := resp.GetMsgIds()[msg.TargetUid]
	msgId := ""
	if len(msgIds.GetItems()) > 0 {
		msgId = msgIds.GetItems()[0]
	}
	return msgId, nil
}

func (a *SystemNotifyAdapterImpl) NotifyReplyMsg(ctx context.Context, msg *vo.SystemMessage) (string, error) {
	resp, err := a.notifyCli.NotifyReplyMsg(ctx,
		&v1.NotifyReplyMsgRequest{
			Contents: []*v1.ReplyMsgContent{
				{
					Uid:       msg.Uid,
					TargetUid: msg.TargetUid,
					Content:   msg.Content,
				},
			},
		})
	if err != nil {
		return "", xerror.Wrap(err)
	}

	msgIds := resp.GetMsgIds()[msg.TargetUid]
	msgId := ""
	if len(msgIds.GetItems()) > 0 {
		msgId = msgIds.GetItems()[0]
	}
	return msgId, nil
}

func (a *SystemNotifyAdapterImpl) NotifyMentionMsg(
	ctx context.Context,
	msgs []*vo.SystemMessage,
) (map[int64][]string, error) {
	mentions := []*v1.MentionMsgContent{}
	for _, msg := range msgs {
		mentions = append(mentions, &v1.MentionMsgContent{
			Uid:       msg.Uid,
			TargetUid: msg.TargetUid,
			Content:   msg.Content,
		})
	}
	resp, err := a.notifyCli.NotifyMentionMsg(ctx,
		&v1.NotifyMentionMsgRequest{
			Mentions: mentions,
		})
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	msgIds := make(map[int64][]string)
	for uid, msg := range resp.GetMsgIds() {
		msgIds[uid] = msg.GetItems()
	}

	return msgIds, nil
}

func (a *SystemNotifyAdapterImpl) ListMentionMsg(
	ctx context.Context,
	uid int64,
	cursor string,
	count int32,
) (*vo.ListMsgResult, error) {
	resp, err := a.chatCli.ListSystemMentionMsg(ctx,
		&v1.ListSystemMentionMsgRequest{
			RecvUid: uid,
			Cursor:  cursor,
			Count:   count,
		})
	if err != nil {
		return nil, xerror.Wrap(err).WithCtx(ctx)
	}

	return convert.ListMsgResultFromPb(resp.GetMessages(), resp.GetChatId(), resp.GetHasMore()), nil
}

func (a *SystemNotifyAdapterImpl) ListReplyMsg(
	ctx context.Context,
	uid int64,
	cursor string,
	count int32,
) (*vo.ListMsgResult, error) {
	resp, err := a.chatCli.ListSystemReplyMsg(ctx,
		&v1.ListSystemReplyMsgRequest{
			RecvUid: uid,
			Cursor:  cursor,
			Count:   count,
		})
	if err != nil {
		return nil, xerror.Wrap(err).WithCtx(ctx)
	}

	return convert.ListMsgResultFromPb(resp.GetMessages(), resp.GetChatId(), resp.GetHasMore()), nil
}

func (a *SystemNotifyAdapterImpl) ListLikesMsg(
	ctx context.Context,
	uid int64,
	cursor string,
	count int32,
) (*vo.ListMsgResult, error) {
	resp, err := a.chatCli.ListSystemLikesMsg(ctx,
		&v1.ListSystemLikesMsgRequest{
			RecvUid: uid,
			Cursor:  cursor,
			Count:   count,
		})
	if err != nil {
		return nil, xerror.Wrap(err).WithCtx(ctx)
	}

	return convert.ListMsgResultFromPb(resp.GetMessages(), resp.GetChatId(), resp.GetHasMore()), nil
}

func (a *SystemNotifyAdapterImpl) GetChatUnread(
	ctx context.Context, uid int64,
) (*entity.ChatsUnreadCount, error) {
	resp, err := a.chatCli.GetAllChatsUnread(ctx,
		&v1.GetAllChatsUnreadRequest{
			Uid: uid,
		})
	if err != nil {
		return nil, xerror.Wrap(err).WithCtx(ctx)
	}

	return &entity.ChatsUnreadCount{
		MentionUnread: convert.ChatUnreadFromPb(resp.GetMentionUnread()),
		NoticeUnread:  convert.ChatUnreadFromPb(resp.GetNoticeUnread()),
		LikesUnread:   convert.ChatUnreadFromPb(resp.GetLikesUnread()),
		ReplyUnread:   convert.ChatUnreadFromPb(resp.GetReplyUnread()),
	}, nil
}

func (a *SystemNotifyAdapterImpl) ClearChatUnread(ctx context.Context, uid int64, chatId string) error {
	_, err := a.chatCli.ClearChatUnread(ctx,
		&v1.ClearChatUnreadRequest{
			Uid:    uid,
			ChatId: chatId,
		})
	if err != nil {
		return xerror.Wrap(err).WithCtx(ctx)
	}
	return nil
}

func (a *SystemNotifyAdapterImpl) DeleteMsg(ctx context.Context, uid int64, msgId string) error {
	_, err := a.chatCli.DeleteMsg(ctx,
		&v1.DeleteMsgRequest{
			Uid:   uid,
			MsgId: msgId,
		})
	if err != nil {
		return xerror.Wrap(err).WithCtx(ctx)
	}
	return nil
}
