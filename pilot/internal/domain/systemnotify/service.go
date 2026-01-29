package systemnotify

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	mentionvo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/mention/vo"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/common/pushcenter"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/entity"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/repository"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/systemnotify/vo"
)

type DomainService struct {
	adapter repository.SystemNotifyAdapter
}

func NewDomainService(adapter repository.SystemNotifyAdapter) *DomainService {
	return &DomainService{adapter: adapter}
}

// 通知用户笔记收到点赞
func (s *DomainService) NotifyUserLikesOnNote(
	ctx context.Context,
	uid, recvUid int64,
	req *vo.NotifyLikesOnNoteParam,
) error {
	msg := vo.NewLikesOnNoteMessage(uid, recvUid, req)
	content, err := json.Marshal(msg)
	if err != nil {
		return xerror.Wrapf(err, "json marshal likes on note msg failed").WithCtx(ctx)
	}

	return s.notifyLikesAndPush(ctx, uid, recvUid, content)
}

// 通知用户评论收到点赞
func (s *DomainService) NotifyUserLikesOnComment(
	ctx context.Context,
	uid, recvUid int64,
	req *vo.NotifyLikesOnCommentParam,
) error {
	msg := vo.NewLikesOnCommentMessage(uid, recvUid, req)
	content, err := json.Marshal(msg)
	if err != nil {
		return xerror.Wrapf(err, "json marshal likes on comment msg failed").WithCtx(ctx)
	}

	return s.notifyLikesAndPush(ctx, uid, recvUid, content)
}

func (s *DomainService) notifyLikesAndPush(
	ctx context.Context,
	uid, recvUid int64,
	content []byte,
) error {
	msgId, err := s.adapter.NotifyLikesMsg(ctx, &vo.SystemMessage{
		Uid:       uid,
		TargetUid: recvUid,
		Content:   content,
	})
	if err != nil {
		return xerror.Wrapf(err, "notify likes msg failed").WithCtx(ctx)
	}

	if msgId != "" {
		if err := pushcenter.NotifySystemMsg(ctx, recvUid); err != nil {
			return xerror.Wrapf(err, "push likes notification failed").WithExtra("recv_uid", recvUid).WithCtx(ctx)
		}
	}

	return nil
}

// 通知用户被回复了
func (s *DomainService) NotifyUserReply(
	ctx context.Context,
	req *vo.NotifyUserReplyParam,
) error {
	content, err := json.Marshal(req)
	if err != nil {
		return xerror.Wrapf(err, "json marshal reply req failed").WithCtx(ctx)
	}

	msgId, err := s.adapter.NotifyReplyMsg(ctx, &vo.SystemMessage{
		Uid:       req.SrcUid,
		TargetUid: req.RecvUid,
		Content:   content,
	})
	if err != nil {
		return xerror.Wrapf(err, "notify reply msg failed").WithCtx(ctx)
	}

	if msgId != "" {
		if err := pushcenter.NotifySystemMsg(ctx, req.RecvUid); err != nil {
			xlog.Msg("push reply notification failed").Extras("recv_uid", req.RecvUid).Errorx(ctx)
			return xerror.Wrapf(err, "push reply notification failed").WithCtx(ctx)
		}
	}

	return nil
}

// 同一份笔记@多个人通知
func (s *DomainService) NotifyAtUsersOnNote(
	ctx context.Context,
	req *vo.NotifyAtUsersOnNoteParam,
) error {
	if len(req.TargetUsers) == 0 {
		return nil
	}

	msgs := s.buildMentionMsgs(req.Uid, req.TargetUsers, &mentionMsgContent{
		NotifyAtUsersOnNoteParamContent: req.Content,
		Receivers:                       toAtUserList(req.TargetUsers),
		Loc:                             vo.NotifyMsgOnNote,
	})

	return s.notifyMentionAndPush(ctx, req.Uid, msgs)
}

// 同一条评论@多个人通知
func (s *DomainService) NotifyAtUsersOnComment(
	ctx context.Context,
	req *vo.NotifyAtUsersOnCommentParam,
) error {
	if len(req.TargetUsers) == 0 {
		return nil
	}

	msgs := s.buildMentionMsgs(req.Uid, req.TargetUsers, &mentionMsgContent{
		NotifyAtUsersOnCommentParamContent: req.Content,
		Receivers:                          toAtUserList(req.TargetUsers),
		Loc:                                vo.NotifyMsgOnComment,
	})

	return s.notifyMentionAndPush(ctx, req.Uid, msgs)
}

func (s *DomainService) buildMentionMsgs(
	uid int64,
	targets []*mentionvo.AtUser,
	content *mentionMsgContent,
) []*vo.SystemMessage {
	contentData, _ := json.Marshal(content)
	msgs := make([]*vo.SystemMessage, 0, len(targets))
	for _, user := range targets {
		msgs = append(msgs, &vo.SystemMessage{
			Uid:       uid,
			TargetUid: user.Uid,
			Content:   contentData,
		})
	}
	return msgs
}

func (s *DomainService) notifyMentionAndPush(
	ctx context.Context,
	uid int64,
	msgs []*vo.SystemMessage,
) error {
	resp, err := s.adapter.NotifyMentionMsg(ctx, msgs)
	if err != nil {
		return xerror.Wrapf(err, "notify mention msg failed").WithExtra("uid", uid).WithCtx(ctx)
	}

	recvUids := make([]int64, 0, len(resp))
	for recvUid := range resp {
		recvUids = append(recvUids, recvUid)
	}

	if err := pushcenter.BatchNotifySystemMsg(ctx, recvUids); err != nil {
		xlog.Msg("push mention notification failed").Err(err).Extras("recv_uids", recvUids).Errorx(ctx)
		return xerror.Wrapf(err, "push mention notification failed").WithCtx(ctx)
	}

	return nil
}

// 获取@消息列表
func (s *DomainService) ListMentionMsg(
	ctx context.Context,
	uid int64,
	cursor string,
	count int32,
) (*vo.ListMsgResult, error) {
	return s.adapter.ListMentionMsg(ctx, uid, cursor, count)
}

// 获取回复消息列表
func (s *DomainService) ListReplyMsg(
	ctx context.Context,
	uid int64,
	cursor string,
	count int32,
) (*vo.ListMsgResult, error) {
	return s.adapter.ListReplyMsg(ctx, uid, cursor, count)
}

// 获取点赞消息列表
func (s *DomainService) ListLikesMsg(
	ctx context.Context,
	uid int64, cursor string, count int32,
) (*vo.ListMsgResult, error) {
	return s.adapter.ListLikesMsg(ctx, uid, cursor, count)
}

// 清除系统会话已读
func (s *DomainService) ClearChatUnread(
	ctx context.Context, uid int64, chatId string,
) error {
	return s.adapter.ClearChatUnread(ctx, uid, chatId)
}

// 获取系统会话的未读数
func (s *DomainService) GetChatUnread(
	ctx context.Context, uid int64,
) (*entity.ChatsUnreadCount, error) {
	return s.adapter.GetChatUnread(ctx, uid)
}

// 删除系统消息
func (s *DomainService) DeleteSysMsg(
	ctx context.Context, uid int64, msgId string,
) error {
	if uid == 0 || msgId == "" {
		return xerror.Wrapf(xerror.ErrArgs, "invalid params").
			WithExtras("uid", uid, "msg_id", msgId).WithCtx(ctx)
	}
	return s.adapter.DeleteMsg(ctx, uid, msgId)
}

type mentionMsgContent struct {
	*vo.NotifyAtUsersOnNoteParamContent    `json:"note_content,omitempty"`
	*vo.NotifyAtUsersOnCommentParamContent `json:"comment_content,omitempty"`
	Receivers                              []*mentionvo.AtUser  `json:"receivers"`
	Loc                                    vo.NotifyMsgLocation `json:"loc"`
}

func toAtUserList(users []*mentionvo.AtUser) []*mentionvo.AtUser {
	result := make([]*mentionvo.AtUser, 0, len(users))
	for _, u := range users {
		result = append(result, &mentionvo.AtUser{Uid: u.Uid, Nickname: u.Nickname})
	}
	return result
}
