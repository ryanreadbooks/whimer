package systemnotify

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"

	v1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
)

type SystemNotifyAdapterImpl struct {
	cli v1.NotificationServiceClient
}

var _ repository.SystemNotifyAdapter = (*SystemNotifyAdapterImpl)(nil)

func NewSystemNotifyAdapterImpl(cli v1.NotificationServiceClient) *SystemNotifyAdapterImpl {
	return &SystemNotifyAdapterImpl{
		cli: cli,
	}
}

func (a *SystemNotifyAdapterImpl) NotifyLikesMsg(ctx context.Context, msg *vo.SystemMessage) (string, error) {
	resp, err := a.cli.NotifyLikesMsg(ctx, &v1.NotifyLikesMsgRequest{
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
	resp, err := a.cli.NotifyReplyMsg(ctx, &v1.NotifyReplyMsgRequest{
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

func (a *SystemNotifyAdapterImpl) NotifyMentionMsg(ctx context.Context, msgs []*vo.SystemMessage) (map[int64][]string, error) {
	mentions := []*v1.MentionMsgContent{}
	for _, msg := range msgs {
		mentions = append(mentions, &v1.MentionMsgContent{
			Uid:       msg.Uid,
			TargetUid: msg.TargetUid,
			Content:   msg.Content,
		})
	}
	resp, err := a.cli.NotifyMentionMsg(ctx, &v1.NotifyMentionMsgRequest{
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
