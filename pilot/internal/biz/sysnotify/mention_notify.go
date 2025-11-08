package sysnotify

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	sysnotifyv1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/common/pushcenter"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
)

type NotifyAtUsersOnNoteReq struct {
	Uid         int64                          `json:"uid"`
	TargetUsers []*notev1.NoteAtUser           `json:"target_users"`
	Content     *NotifyAtUsersOnNoteReqContent `json:"content"`
}

type NotifyAtUsersOnNoteReqContent struct {
	SourceUid int64         `json:"src_uid"` // trigger uid
	NoteDesc  string        `json:"desc"`
	NoteId    imodel.NoteId `json:"id"` // 笔记id
}

func mapToMentionReceiver(user imodel.IAtUser) *model.MentionedRecvUser {
	return &model.MentionedRecvUser{Nickname: user.GetNickname(), Uid: user.GetUid()}
}

// 消息内容 反序列化的时候用这个进行反序列化
type notifyAtUserReqContent struct {
	*NotifyAtUsersOnNoteReqContent    `json:"note_content,omitempty"`
	*NotifyAtUsersOnCommentReqContent `json:"comment_content,omitempty"`

	Receivers []*model.MentionedRecvUser `json:"receivers"`
	Loc       model.NotifyMsgLocation    `json:"loc"` // @人的位置
}

// 同一份笔记@多个人通知
func (b *Biz) NotifyAtUsersOnNote(ctx context.Context, req *NotifyAtUsersOnNoteReq) error {
	if len(req.TargetUsers) == 0 {
		return nil
	}
	mRecvs := make([]*model.MentionedRecvUser, 0, len(req.TargetUsers))
	for _, user := range req.TargetUsers {
		mRecvs = append(mRecvs, mapToMentionReceiver(user))
	}

	mentions := make([]*sysnotifyv1.MentionMsgContent, 0, len(req.TargetUsers))
	for _, user := range req.TargetUsers {
		contentData, _ := json.Marshal(&notifyAtUserReqContent{
			NotifyAtUsersOnNoteReqContent: req.Content,
			Receivers:                     mRecvs,
			Loc:                           model.NotifyMsgOnNote,
		})
		mentions = append(mentions, &sysnotifyv1.MentionMsgContent{
			Uid:       req.Uid,
			TargetUid: user.Uid,
			Content:   contentData,
		})
	}

	resp, err := dep.SystemNotifier().NotifyMentionMsg(ctx, &sysnotifyv1.NotifyMentionMsgRequest{
		Mentions: mentions,
	})
	if err != nil {
		return xerror.Wrapf(err, "system notifier mention msg failed").
			WithExtras("uid", req.Uid).WithCtx(ctx)
	}

	// 通知用户拉取最新的系统消息
	recvUids := make([]int64, 0, len(resp.GetMsgIds()))
	for uid := range resp.GetMsgIds() {
		recvUids = append(recvUids, uid)
	}
	if err := pushcenter.BatchNotifySystemMsg(ctx, recvUids); err != nil {
		xlog.Msg("sysnotify biz push mention on note notification failed").
			Err(err).Extras("recv_uids", recvUids).Errorx(ctx)
		return xerror.Wrapf(err, "push mention notification failed").WithCtx(ctx)
	}

	return nil
}

type NotifyAtUsersOnCommentReq struct {
	Uid         int64                             `json:"uid"`          // 谁@
	TargetUsers []imodel.AtUser                   `json:"target_users"` // 谁被@
	Content     *NotifyAtUsersOnCommentReqContent `json:"content"`
}

type NotifyAtUsersOnCommentReqContent struct {
	SourceUid int64         `json:"src_uid"`    // 评论发布者uid
	Comment   string        `json:"comment"`    // 评论内容
	NoteId    imodel.NoteId `json:"note_id"`    // 评论归属笔记id
	CommentId int64         `json:"comment_id"` // 评论id
	RootId    int64         `json:"root_id"`    // 根评论id
	ParentId  int64         `json:"parent_id"`  // 父评论id
}

// 同一条评论@多个人通知
func (b *Biz) NotifyAtUsersOnComment(ctx context.Context, req *NotifyAtUsersOnCommentReq) error {
	if len(req.TargetUsers) == 0 {
		return nil
	}

	mRecvs := make([]*model.MentionedRecvUser, 0, len(req.TargetUsers))
	for _, user := range req.TargetUsers {
		mRecvs = append(mRecvs, mapToMentionReceiver(&user))
	}

	mentions := make([]*sysnotifyv1.MentionMsgContent, 0, len(req.TargetUsers))
	for _, ated := range req.TargetUsers {
		contentData, _ := json.Marshal(&notifyAtUserReqContent{
			NotifyAtUsersOnCommentReqContent: req.Content,
			Receivers:                        mRecvs,
			Loc:                              model.NotifyMsgOnComment,
		})
		mentions = append(mentions, &sysnotifyv1.MentionMsgContent{
			Content:   contentData,
			Uid:       req.Uid,
			TargetUid: ated.Uid,
		})
	}

	resp, err := dep.SystemNotifier().NotifyMentionMsg(ctx, &sysnotifyv1.NotifyMentionMsgRequest{
		Mentions: mentions,
	})
	if err != nil {
		return xerror.Wrapf(err, "system notifier mention msg failed").
			WithExtras("uid", req.Uid).WithCtx(ctx)
	}

	// 通知用户拉取最新的系统消息
	recvUids := make([]int64, 0, len(resp.GetMsgIds()))
	for uid := range resp.GetMsgIds() {
		recvUids = append(recvUids, uid)
	}
	if err := pushcenter.BatchNotifySystemMsg(ctx, recvUids); err != nil {
		xlog.Msg("sysnotify biz push mention on note notification failed").
			Err(err).Extras("recv_uids", recvUids).Errorx(ctx)
	}

	return nil
}
