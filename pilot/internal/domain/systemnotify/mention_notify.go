package systemnotify

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/pilot/internal/biz/common/pushcenter"
	mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

// 消息内容 反序列化的时候用这个进行反序列化
type notifyAtUserReqContent struct {
	*vo.NotifyAtUsersOnNoteParamContent    `json:"note_content,omitempty"`
	*vo.NotifyAtUsersOnCommentParamContent `json:"comment_content,omitempty"`

	Receivers []*mentionvo.AtUser  `json:"receivers"`
	Loc       vo.NotifyMsgLocation `json:"loc"` // @人的位置
}

// 同一份笔记@多个人通知
func (b *DomainService) NotifyAtUsersOnNote(ctx context.Context, req *vo.NotifyAtUsersOnNoteParam) error {
	if len(req.TargetUsers) == 0 {
		return nil
	}

	mRecvs := make([]*mentionvo.AtUser, 0, len(req.TargetUsers))
	for _, user := range req.TargetUsers {
		mRecvs = append(mRecvs, &mentionvo.AtUser{
			Uid:      user.Uid,
			Nickname: user.Nickname,
		})
	}

	msgs := make([]*vo.SystemMessage, 0, len(req.TargetUsers))
	for _, user := range req.TargetUsers {
		contentData, _ := json.Marshal(&notifyAtUserReqContent{
			NotifyAtUsersOnNoteParamContent: req.Content,

			Receivers: mRecvs,
			Loc:       vo.NotifyMsgOnNote,
		})

		msgs = append(msgs, &vo.SystemMessage{
			Uid:       req.Uid,
			TargetUid: user.Uid,
			Content:   contentData,
		})
	}

	resp, err := b.systemNotifyAdapter.NotifyMentionMsg(ctx, msgs)
	if err != nil {
		return xerror.Wrapf(err, "system notifier mention msg failed").
			WithExtras("uid", req.Uid).WithCtx(ctx)
	}

	// 通知用户拉取最新的系统消息
	recvUids := make([]int64, 0, len(resp))
	for uid := range resp {
		recvUids = append(recvUids, uid)
	}
	if err := pushcenter.BatchNotifySystemMsg(ctx, recvUids); err != nil {
		xlog.Msg("sysnotify biz push mention on note notification failed").
			Err(err).Extras("recv_uids", recvUids).Errorx(ctx)
		return xerror.Wrapf(err, "push mention notification failed").WithCtx(ctx)
	}

	return nil
}

// 同一条评论@多个人通知
func (b *DomainService) NotifyAtUsersOnComment(ctx context.Context, req *vo.NotifyAtUsersOnCommentParam) error {
	if len(req.TargetUsers) == 0 {
		return nil
	}

	mRecvs := make([]*mentionvo.AtUser, 0, len(req.TargetUsers))
	for _, user := range req.TargetUsers {
		mRecvs = append(mRecvs, &mentionvo.AtUser{Uid: user.Uid, Nickname: user.Nickname})
	}

	msgs := make([]*vo.SystemMessage, 0, len(req.TargetUsers))
	for _, ated := range req.TargetUsers {
		contentData, _ := json.Marshal(&notifyAtUserReqContent{
			NotifyAtUsersOnCommentParamContent: req.Content,

			Receivers: mRecvs,
			Loc:       vo.NotifyMsgOnComment,
		})
		msgs = append(msgs, &vo.SystemMessage{
			Uid:       req.Uid,
			TargetUid: ated.Uid,
			Content:   contentData,
		})
	}

	resp, err := b.systemNotifyAdapter.NotifyMentionMsg(ctx, msgs)
	if err != nil {
		return xerror.Wrapf(err, "system notifier mention msg failed").
			WithExtras("uid", req.Uid).WithCtx(ctx)
	}

	// 通知用户拉取最新的系统消息
	recvUids := make([]int64, 0, len(resp))
	for uid := range resp {
		recvUids = append(recvUids, uid)
	}
	if err := pushcenter.BatchNotifySystemMsg(ctx, recvUids); err != nil {
		xlog.Msg("sysnotify biz push mention on note notification failed").
			Err(err).Extras("recv_uids", recvUids).Errorx(ctx)
	}

	return nil
}
