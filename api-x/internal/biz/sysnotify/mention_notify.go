package notification

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/api-x/internal/biz/sysnotify/model"
	usermodel "github.com/ryanreadbooks/whimer/api-x/internal/biz/user/model"
	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	imodel "github.com/ryanreadbooks/whimer/api-x/internal/model"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	sysnotifyv1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
)

type Biz struct {
}

func NewBiz() *Biz { return &Biz{} }

type NotifyAtUsersOnNoteReq struct {
	Uid         int64                          `json:"uid"`
	TargetUsers []*notev1.NoteAtUser           `json:"target_users"`
	Content     *NotifyAtUsersOnNoteReqContent `json:"content"`
}

type NotifyAtUsersOnNoteReqContent struct {
	SourceUid int64         `json:"src_uid"` // trigger uid
	NoteDesc  string        `json:"desc"`
	NoteId    imodel.NoteId `json:"id"`
}

// 消息内容 反序列化的时候用这个进行反序列化
type notifyAtUserReqContent struct {
	*NotifyAtUsersOnNoteReqContent    `json:"note_content,omitempty"`
	*NotifyAtUsersOnCommentReqContent `json:"comment_content,omitempty"`

	RecvUid      int64                 `json:"recv_uid"`
	RecvNickname string                `json:"recv_nickname"`
	Loc          model.MentionLocation `json:"loc"` // @人的位置
}

// 同一份笔记@多个人通知
func (b *Biz) NotifyAtUsersOnNote(ctx context.Context, req *NotifyAtUsersOnNoteReq) error {
	if len(req.TargetUsers) == 0 {
		return nil
	}

	mentions := make([]*sysnotifyv1.MentionMsgContent, 0, len(req.TargetUsers))
	for _, user := range req.TargetUsers {
		contentData, _ := json.Marshal(&notifyAtUserReqContent{
			NotifyAtUsersOnNoteReqContent: req.Content,
			RecvUid:                       user.Uid,
			RecvNickname:                  user.Nickname,
			Loc:                           model.MentionOnNote,
		})
		mentions = append(mentions, &sysnotifyv1.MentionMsgContent{
			Uid:       req.Uid,
			TargetUid: user.Uid,
			Content:   contentData,
		})
	}

	_, err := infra.SystemNotifier().NotifyMentionMsg(ctx, &sysnotifyv1.NotifyMentionMsgRequest{
		Mentions: mentions,
	})
	if err != nil {
		return xerror.Wrapf(err, "system notifier mention msg failed").
			WithExtras("uid", req.Uid).WithCtx(ctx)
	}

	return nil
}

type NotifyAtUsersOnCommentReq struct {
	Uid         int64                             `json:"uid"`
	TargetUsers []*usermodel.User                 `json:"target_users"` //TODO change user type
	Content     *NotifyAtUsersOnCommentReqContent `json:"content"`
}

type NotifyAtUsersOnCommentReqContent struct {
	SourceUid int64         `json:"src_uid"` // trigger uid
	RecvUid   int64         `json:"recv_uid"`
	Comment   string        `json:"comment"`
	NoteId    imodel.NoteId `json:"note_id"`
	CommentId int64         `json:"comment_id"`
}

func (b *Biz) NotifyAtUsersOnComment(ctx context.Context, req *NotifyAtUsersOnCommentReq) error {
	if len(req.TargetUsers) == 0 {
		return nil
	}

	mentions := make([]*sysnotifyv1.MentionMsgContent, 0, len(req.TargetUsers))
	for _, user := range req.TargetUsers {
		contentData, _ := json.Marshal(&notifyAtUserReqContent{
			NotifyAtUsersOnCommentReqContent: req.Content,
			RecvUid:                          user.Uid,
			Loc:                              model.MentionOnComment,
		})
		mentions = append(mentions, &sysnotifyv1.MentionMsgContent{
			Content:   contentData,
			Uid:       req.Uid,
			TargetUid: user.Uid,
		})
	}

	_, err := infra.SystemNotifier().NotifyMentionMsg(ctx, &sysnotifyv1.NotifyMentionMsgRequest{
		Mentions: mentions,
	})
	if err != nil {
		return xerror.Wrapf(err, "system notifier mention msg failed").
			WithExtras("uid", req.Uid).WithCtx(ctx)
	}

	return nil
}
