package sysnotify

import (
	"context"
	"encoding/json"

	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/recovery"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xslice"
	sysnotifyv1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
	systemv1 "github.com/ryanreadbooks/whimer/msger/api/system/v1"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dao/kafka"
	sysmsgkfkdao "github.com/ryanreadbooks/whimer/pilot/internal/infra/dao/kafka/sysmsg"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	imodel "github.com/ryanreadbooks/whimer/pilot/internal/model"
	"golang.org/x/sync/errgroup"

	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

type ListUserMentionMsgResp struct {
	Msgs    []*model.MentionedMsg
	ChatId  string
	HasNext bool
}

type ListUserReplyMsgResp struct {
	Msgs    []*model.ReplyMsg
	ChatId  string
	HasNext bool
}

// 获取用户uid的被@消息
func (b *Biz) ListUserMentionMsg(ctx context.Context, uid int64, cursor string, count int32) (*ListUserMentionMsgResp, error) {
	result := &ListUserMentionMsgResp{}
	resp, err := dep.SystemChatter().ListSystemMentionMsg(ctx, &sysnotifyv1.ListSystemMentionMsgRequest{
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
		mgid, err := uuid.ParseString(msg.Id)
		if err != nil {
			// should not be err
			xlog.Msg("parse mention msg id failed, it should be successful").
				Err(err).
				Extras("msgid", msg.Id).
				Errorx(ctx)
			continue
		}
		mm := model.MentionedMsg{
			Id:     msg.Id,
			SendAt: mgid.UnixSec(),
		}

		if msg.Status != sysnotifyv1.SystemMsgStatus_MsgStatus_Revoked {
			// 不是撤回的消息可以直接反序列化
			var v notifyAtUserReqContent
			err = json.Unmarshal(msg.Content, &v)
			if err != nil {
				xlog.Msg("unmarshal mention msg content failed").Err(err).Errorx(ctx)
				continue
			}

			var (
				loc       model.MentionLocation
				sourceUid int64
				noteId    imodel.NoteId = 0
				content   string
				commentId int64 = 0
			)

			if v.NotifyAtUsersOnNoteReqContent != nil {
				loc = model.MentionOnNote
				sourceUid = v.NotifyAtUsersOnNoteReqContent.SourceUid
				noteId = v.NotifyAtUsersOnNoteReqContent.NoteId
				content = v.NotifyAtUsersOnNoteReqContent.NoteDesc
			} else if v.NotifyAtUsersOnCommentReqContent != nil {
				loc = model.MentionOnComment
				sourceUid = v.NotifyAtUsersOnCommentReqContent.SourceUid
				content = v.NotifyAtUsersOnCommentReqContent.Comment
				commentId = v.NotifyAtUsersOnCommentReqContent.CommentId
				noteId = v.NotifyAtUsersOnCommentReqContent.NoteId
			}

			mm.Type = loc
			mm.Uid = sourceUid
			mm.RecvUsers = v.Receivers
			mm.NoteId = noteId
			mm.CommentId = commentId
			mm.Content = content
			mm.Status = model.MsgStatusNormal
		} else {
			mm.Status = model.MsgStatusRevoked
		}
		mentionMsgs = append(mentionMsgs, &mm)
	}

	if err := b.lazyCheckMentionSource(ctx, mentionMsgs); err != nil {
		return nil, xerror.Wrapf(err, "sysmsg biz lazy check mention source failed")
	}

	result.Msgs = mentionMsgs
	result.HasNext = resp.HasMore
	result.ChatId = resp.ChatId

	return result, nil
}

// 获取用户uid被回复的消息
func (b *Biz) ListUserReplyMsg(ctx context.Context, uid int64, cursor string, count int32) (*ListUserReplyMsgResp, error) {
	resp, err := dep.SystemChatter().ListSystemReplyMsg(ctx, &systemv1.ListSystemReplyMsgRequest{
		RecvUid: uid,
		Cursor:  cursor,
		Count:   count,
	})
	if err != nil {
		return nil, xerror.Wrapf(err, "system chatter list system reply msg failed").
			WithExtras("uid", uid, "cursor", cursor, "count", count).
			WithCtx(ctx)
	}

	var (
		mLen = len(resp.GetMessages())
	)

	replyMsgs := make([]*model.ReplyMsg, 0, mLen)
	for _, msg := range resp.GetMessages() {
		mgid, err := uuid.ParseString(msg.Id)
		if err != nil {
			// should not be err
			xlog.Msg("parse mention msg id failed, it should be successful").
				Err(err).
				Extras("msgid", msg.Id).
				Errorx(ctx)
			continue
		}

		rm := model.ReplyMsg{
			Id:      msg.Id,
			SendAt:  mgid.UnixSec(),
			RecvUid: uid,
		}

		if msg.Status != sysnotifyv1.SystemMsgStatus_MsgStatus_Revoked {
			// 解析msg content
			var content NotifyUserReplyReq
			err := json.Unmarshal(msg.Content, &content)
			if err != nil {
				xlog.Msg("sysmsg biz json unmarshal content failed").Extras("msg_id", msg.Id).Errorx(ctx)
				continue
			}

			var commentContent imodel.CommentContent
			err = json.Unmarshal(content.Content, &commentContent)
			if err != nil {
				xlog.Msg("sysmsg biz json unmarshal comment failed").Extras("msg_id", msg.Id).Errorx(ctx)
				continue
			}

			rm.NoteId = content.NoteId
			rm.Uid = content.SrcUid
			rm.Status = model.MsgStatusNormal
			rm.Content = commentContent.Text
			rm.Type = content.Loc
			rm.TargetComment = content.TargetComment
			rm.TriggerComment = content.TriggerComment
			if len(commentContent.AtUsers) > 0 {
				rm.Ext = &model.ReplyMsgExt{
					AtUsers: commentContent.AtUsers,
				}
			}
		} else {
			rm.Status = model.MsgStatusRevoked
		}

		replyMsgs = append(replyMsgs, &rm)
	}

	if err := b.lazyCheckReplySource(ctx, replyMsgs); err != nil {
		return nil, xerror.Wrapf(err, "sysmsg biz check reply source failed").WithCtx(ctx)
	}

	return &ListUserReplyMsgResp{
		Msgs:    replyMsgs,
		ChatId:  resp.ChatId,
		HasNext: resp.HasMore,
	}, nil
}

// 清除系统会话已读
func (b *Biz) ClearChatUnread(ctx context.Context, uid int64, chatId string) error {
	_, err := dep.SystemChatter().ClearChatUnread(ctx, &systemv1.ClearChatUnreadRequest{
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
	resp, err := dep.SystemChatter().GetAllChatsUnread(ctx, &systemv1.GetAllChatsUnreadRequest{
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

// 检查原始数据是否还存在 不存在需要屏蔽掉对应消息
func (b *Biz) lazyCheckMentionSource(ctx context.Context, msgs []*model.MentionedMsg) error {
	var (
		uid        = metadata.Uid(ctx)
		numMsgs    = len(msgs)
		noteIds    = make([]int64, 0, numMsgs)
		commentIds = make([]int64, 0, numMsgs)
	)

	for _, msg := range msgs {
		noteId := int64(msg.NoteId)
		switch msg.Type {
		case model.MentionOnComment:
			noteIds = append(noteIds, noteId)
			commentIds = append(commentIds, msg.CommentId)
		case model.MentionOnNote:
			noteIds = append(noteIds, noteId)
		}
	}

	noteExistence, commentExistence, err := b.checkSourcesExistence(ctx, noteIds, commentIds)
	if err != nil {
		return xerror.Wrapf(err, "sysmsg check source failed").WithCtx(ctx)
	}

	pendingMsgIds := make([]string, 0, numMsgs)

	// noteExistence
	for _, msg := range msgs {
		noteId := int64(msg.NoteId)
		switch msg.Type {
		case model.MentionOnComment:
			noteOk, _ := noteExistence[noteId]
			commentOk, _ := commentExistence[msg.CommentId]
			if !noteOk || !commentOk {
				msg.DoNotReveal()
				pendingMsgIds = append(pendingMsgIds, msg.Id)
			}
		case model.MentionOnNote:
			noteOk, _ := noteExistence[noteId]
			if !noteOk {
				msg.DoNotReveal()
				pendingMsgIds = append(pendingMsgIds, msg.Id)
			}
		}
	}

	xlog.Msgf("sysmsg check pending mention msgids length = %d", len(pendingMsgIds)).Debugx(ctx)

	// batch delete system msgs for the same uid (by kafka)
	if len(pendingMsgIds) > 0 {
		deletions := make([]*sysmsgkfkdao.DeletionEvent, 0, len(pendingMsgIds))
		for _, msgId := range pendingMsgIds {
			deletions = append(deletions, &sysmsgkfkdao.DeletionEvent{
				Uid:   uid,
				MsgId: msgId,
			})
		}

		if err := kafka.Dao().SysMsgEventProducer.AsyncPutDeletion(ctx, deletions); err != nil {
			xlog.Msg("sysmsg biz async put deletion failed").Err(err).Extras("args", deletions).Errorx(ctx)
		}
	}

	return nil
}

// 检查原始信息是否存在 不存在需要屏蔽消息
func (b *Biz) lazyCheckReplySource(ctx context.Context, msgs []*model.ReplyMsg) error {
	var (
		uid        = metadata.Uid(ctx)
		numMsgs    = len(msgs)
		noteIds    = make([]int64, 0, numMsgs)
		commentIds = make([]int64, 0, numMsgs)
	)
	for _, msg := range msgs {
		switch msg.Type {
		case model.ReplyOnNote:
			noteIds = append(noteIds, int64(msg.NoteId))
		case model.ReplyOnComment:
			noteIds = append(noteIds, int64(msg.NoteId))
			commentIds = append(commentIds, msg.TargetComment)
			commentIds = append(commentIds, msg.TriggerComment)
		}
	}

	noteIds = xslice.Uniq(noteIds)
	commentIds = xslice.Uniq(commentIds)
	noteExistence, commentExistence, err := b.checkSourcesExistence(ctx, noteIds, commentIds)
	if err != nil {
		return xerror.Wrapf(err, "sysmsg biz check source failed").WithCtx(ctx)
	}

	pendingMsgIds := make([]string, 0, numMsgs)

	// noteExistence
	for _, msg := range msgs {
		noteId := int64(msg.NoteId)
		switch msg.Type {
		case model.ReplyOnComment:
			noteOk, _ := noteExistence[noteId]
			commentOk, _ := commentExistence[msg.TriggerComment]
			commentOk2 := commentExistence[msg.TargetComment]
			if !noteOk || !commentOk || !commentOk2 {
				msg.DoNotReveal()
				pendingMsgIds = append(pendingMsgIds, msg.Id)
			}
		case model.ReplyOnNote:
			noteOk, _ := noteExistence[noteId]
			if !noteOk {
				msg.DoNotReveal()
				pendingMsgIds = append(pendingMsgIds, msg.Id)
			}
		}
	}

	xlog.Msgf("sysmsg check pending reply msgids length = %d", len(pendingMsgIds)).Debugx(ctx)

	// batch delete system msgs for the same uid (by kafka)
	if len(pendingMsgIds) > 0 {
		deletions := make([]*sysmsgkfkdao.DeletionEvent, 0, len(pendingMsgIds))
		for _, msgId := range pendingMsgIds {
			deletions = append(deletions, &sysmsgkfkdao.DeletionEvent{
				Uid:   uid,
				MsgId: msgId,
			})
		}

		if err := kafka.Dao().SysMsgEventProducer.AsyncPutDeletion(ctx, deletions); err != nil {
			xlog.Msg("sysmsg biz async put deletion failed").Err(err).Extras("args", deletions).Errorx(ctx)
		}
	}

	return nil
}

func (b *Biz) checkSourcesExistence(ctx context.Context, noteIds, commentIds []int64) (
	noteExistence, commentExistence map[int64]bool, err error) {

	noteExistence = make(map[int64]bool)
	commentExistence = make(map[int64]bool)

	// batch check
	eg, ctx := errgroup.WithContext(ctx)
	if len(noteIds) > 0 {
		eg.Go(recovery.DoV2(func() error {
			resp, err := dep.NoteFeedServer().BatchCheckFeedNoteExist(ctx,
				&notev1.BatchCheckFeedNoteExistRequest{
					NoteIds: noteIds,
				})
			if err != nil {
				return xerror.Wrapf(err, "batch check note failed").WithExtras("note_ids", noteIds)
			}

			noteExistence = resp.GetExistence()

			return nil
		}))
	}

	if len(commentIds) > 0 {
		eg.Go(recovery.DoV2(func() error {
			resp, err := dep.Commenter().BatchCheckCommentExist(ctx,
				&commentv1.BatchCheckCommentExistRequest{
					Ids: commentIds,
				})
			if err != nil {
				return xerror.Wrapf(err, "batch check comment failed").WithExtras("comment_ids", commentIds)
			}

			commentExistence = resp.GetExistence()
			return nil
		}))
	}

	err = eg.Wait()
	if err != nil {
		return nil, nil, xerror.Wrap(err).WithCtx(ctx)
	}

	return noteExistence, commentExistence, nil
}

// 删除系统消息
func (b *Biz) DeleteSysMsg(ctx context.Context, uid int64, msgId string) error {
	if uid == 0 || msgId == "" {
		return xerror.Wrapf(xerror.ErrArgs, "invalid params").WithExtras("uid", uid, "msg_id", msgId).WithCtx(ctx)
	}

	_, err := dep.SystemChatter().DeleteMsg(ctx, &systemv1.DeleteMsgRequest{
		MsgId: msgId,
		Uid:   uid,
	})
	if err != nil {
		return xerror.Wrapf(err, "sysmsg biz failed to delete msg").WithCtx(ctx)
	}

	return nil
}
