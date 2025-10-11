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

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xlog"
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

// 获取用户的被@消息
func (b *Biz) ListUserMentionMsg(ctx context.Context, uid int64, cursor string, count int32) ([]*model.MentionedMsg, bool, error) {
	resp, err := infra.SystemNotifier().ListSystemMentionMsg(ctx, &sysnotifyv1.ListSystemMentionMsgRequest{
		RecvUid: uid,
		Cursor:  cursor,
		Count:   count,
	})
	if err != nil {
		return nil, false, xerror.Wrapf(err, "system notifier list mention msg failed").
			WithExtras("uid", uid, "cursor", cursor, "count", count).
			WithCtx(ctx)
	}

	var (
		mLen = len(resp.GetMessages())
	)

	mentionMsgs := make([]*model.MentionedMsg, 0, mLen)
	for _, msg := range resp.GetMessages() {
		if msg.Status != sysnotifyv1.SystemMsgStatus_SystemMsgStatus_Revoked {
			// 不是撤回的消息可以直接反序列化
			var v notifyAtUserReqContent
			err = json.Unmarshal(msg.Content, &v)
			if err != nil {
				xlog.Msg("unmarshal mention msg content failed").Err(err).Errorx(ctx)
				continue
			}

			mgid, err := uuid.ParseString(msg.Id)
			if err != nil {
				// should not be err
				xlog.Msg("parse mention msg id failed, it should be successful").
					Err(err).
					Extras("msgid", msg.Id).
					Errorx(ctx)
				continue
			}

			var (
				loc       model.MentionLocation
				uid       int64
				noteId    imodel.NoteId = 0
				content   string
				commentId int64 = 0
			)

			if v.NotifyAtUsersOnNoteReqContent != nil {
				loc = model.MentionOnNote
				uid = v.NotifyAtUsersOnNoteReqContent.SourceUid
				noteId = v.NotifyAtUsersOnNoteReqContent.NoteId
				content = v.NotifyAtUsersOnNoteReqContent.NoteDesc
			} else if v.NotifyAtUsersOnCommentReqContent != nil {
				loc = model.MentionOnComment
				uid = v.NotifyAtUsersOnCommentReqContent.SourceUid
				content = v.NotifyAtUsersOnCommentReqContent.Comment
				commentId = v.NotifyAtUsersOnCommentReqContent.CommentId
			}

			mm := model.MentionedMsg{
				Id:     msg.Id,
				SendAt: mgid.UnixSec(),
				Type:   loc,
				Uid:    uid,
				RecvUser: &model.MentionedRecvUser{
					Uid:      v.RecvUid,
					Nickname: v.RecvNickname,
				},
				NoteId:    noteId,
				CommentId: commentId,
				Content:   content,
			}

			mentionMsgs = append(mentionMsgs, &mm)
		}
	}

	return mentionMsgs, resp.HasMore, nil
}
