package notification

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/api-x/internal/biz/sysnotify/model"
	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	imodel "github.com/ryanreadbooks/whimer/api-x/internal/model"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	sysnotifyv1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
	v1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

type ListUserMentionMsgResp struct {
	Msgs    []*model.MentionedMsg
	ChatId  string
	HasNext bool
}

// 获取用户的被@消息
func (b *Biz) ListUserMentionMsg(ctx context.Context, uid int64, cursor string, count int32) (*ListUserMentionMsgResp, error) {
	result := &ListUserMentionMsgResp{}
	resp, err := infra.SystemChatter().ListSystemMentionMsg(ctx, &sysnotifyv1.ListSystemMentionMsgRequest{
		RecvUid: uid,
		Cursor:  cursor,
		Count:   count,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "system chatter list mention msg failed").
			WithExtras("uid", uid, "cursor", cursor, "count", count).
			WithCtx(ctx)
	}

	var (
		mLen = len(resp.GetMessages())
	)

	mentionMsgs := make([]*model.MentionedMsg, 0, mLen)
	for _, msg := range resp.GetMessages() {
		if msg.Status != sysnotifyv1.SystemMsgStatus_MsgStatus_Revoked {
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

	result.Msgs = mentionMsgs
	result.HasNext = resp.HasMore
	result.ChatId = resp.ChatId

	return result, nil
}

// 清除系统会话已读
func (b *Biz) ClearChatUnread(ctx context.Context, uid int64, chatId string) error {
	_, err := infra.SystemChatter().ClearChatUnread(ctx, &v1.ClearChatUnreadRequest{
		Uid:    uid,
		ChatId: chatId,
	})
	if err != nil {
		return xerror.Wrapf(err, "system chatter clear chat unread failed").
			WithExtra("chat_id", chatId).
			WithCtx(ctx)
	}

	return nil
}

// 获取系统会话的未读数
func (b *Biz) GetChatUnread(ctx context.Context, uid int64) (*model.ChatsUnreadCount, error) {
	resp, err := infra.SystemChatter().GetAllChatsUnread(ctx, &v1.GetAllChatsUnreadRequest{
		Uid: uid,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "system chatter get all chats unread failed").
			WithExtra("uid", uid).
			WithCtx(ctx)
	}

	result := &model.ChatsUnreadCount{
		MentionUnread: model.ChatUnreadFromPb(resp.MentionUnread),
		NoticeUnread:  model.ChatUnreadFromPb(resp.NoticeUnread),
		LikesUnread:   model.ChatUnreadFromPb(resp.LikesUnread),
		ReplyUnread:   model.ChatUnreadFromPb(resp.ReplyUnread),
	}

	return result, nil
}
