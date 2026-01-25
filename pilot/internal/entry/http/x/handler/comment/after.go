package comment

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	bizsysnotify "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify"
	sysnotifymodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/core/dep"
	"github.com/ryanreadbooks/whimer/pilot/internal/model"
)

func (h *Handler) checkHasNote(ctx context.Context, noteId int64) error {
	if resp, err := dep.NoteCreatorServer().IsNoteExist(ctx,
		&notev1.IsNoteExistRequest{
			NoteId: noteId,
		}); err != nil {
		return err
	} else {
		if !resp.Exist {
			return xerror.ErrArgs.Msg("笔记不存在")
		}
	}

	return nil
}

// afterNoteCommented 评论发布后的后置处理
func (h *Handler) afterNoteCommented(ctx context.Context, commentId int64, req *PubReq) {
	noteId := req.Oid.String()
	h.syncCommentCountToSearcher(ctx, noteId, 1)
	h.notifyWhenAtUsers(ctx, commentId, req)
	h.asyncNotifyReplyUser(ctx, commentId, req)
	h.appendRecentContacts(ctx, req.AtUsers)
}

func (h *Handler) appendRecentContacts(ctx context.Context, atUsers model.AtUserList) {
	uid := metadata.Uid(ctx)
	h.userApp.AsyncAppendRecentContactsAtUser(ctx, uid, atUsers)
}

func (h *Handler) syncCommentCountToSearcher(ctx context.Context, noteId string, incr int64) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "comment.handler.synces",
		Job: func(ctx context.Context) error {
			err := h.searchBiz.NoteStatSyncer.AddCommentCount(ctx, noteId, incr)
			if err != nil {
				xlog.Msg("note stat add comment count failed").
					Extras("note_id", noteId, "incr", incr).
					Err(err).Errorx(ctx)
				return err
			}

			return err
		},
	})
}

func (h *Handler) notifyWhenAtUsers(ctx context.Context, commentId int64, req *PubReq) {
	uid := metadata.Uid(ctx)

	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "comment.handler.notify_at_users",
		Job: func(ctx context.Context) error {
			err := h.notifyBiz.NotifyAtUsersOnComment(ctx, &bizsysnotify.NotifyAtUsersOnCommentReq{
				Uid:         uid,
				TargetUsers: req.AtUsers,
				Content: &bizsysnotify.NotifyAtUsersOnCommentReqContent{
					SourceUid: uid,
					Comment:   req.Content,
					NoteId:    req.Oid,
					CommentId: commentId,
					RootId:    req.RootId,
					ParentId:  req.ParentId,
				},
			})
			if err != nil {
				xlog.Msg("notify when at users failed").
					Extras("at_users", req.AtUsers, "oid", req.Oid).
					Err(err).Errorx(ctx)
				return err
			}

			return nil
		},
	})
}

func (h *Handler) asyncNotifyReplyUser(ctx context.Context, triggerComment int64, req *PubReq) {
	var (
		uid           = metadata.Uid(ctx)
		loc           sysnotifymodel.NotifyMsgLocation
		targetComment int64
	)

	if req.PubOnOidDirectly() {
		loc = sysnotifymodel.NotifyMsgOnNote
	} else {
		loc = sysnotifymodel.NotifyMsgOnComment
		targetComment = req.ParentId
	}

	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "comment.handler.notify_reply_users",
		Job: func(ctx context.Context) error {
			cmtContent, err := h.commentBiz.GetCommentContent(ctx, triggerComment)
			if err != nil {
				xlog.Msg("notify reply to user failed when get comment content").
					Extras("req", req, "trigger_comment", triggerComment).
					Err(err).
					Errorx(ctx)
				return err
			}

			content, err := json.Marshal(cmtContent)
			if err != nil {
				return xerror.Wrapf(err, "json marshal content failed")
			}

			err = h.notifyBiz.NotifyUserReply(ctx, &bizsysnotify.NotifyUserReplyReq{
				Loc:            loc,
				TriggerComment: triggerComment,
				TargetComment:  targetComment,
				SrcUid:         uid,
				RecvUid:        req.ReplyUid,
				NoteId:         req.Oid,
				Content:        content,
			})
			if err != nil {
				xlog.Msg("notify reply to user failed when notifying").
					Extras("req", req, "trigger_comment", triggerComment).
					Err(err).
					Errorx(ctx)
				return err
			}

			return nil
		},
	})
}

func (h *Handler) asyncNotifyLikeComment(ctx context.Context, commentId int64) {
	uid := metadata.Uid(ctx)

	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name:       "comment.handler.notify_likes",
		LogOnError: true,
		Job: func(ctx context.Context) error {
			item, err := h.commentBiz.GetComment(ctx, commentId)
			if err != nil {
				return xerror.Wrapf(err, "comment biz get comment failed").WithExtras("comment_id", commentId).WithCtx(ctx)
			}

			recvUid := item.Uid // 作者
			if recvUid == 0 {
				return nil
			}

			err = h.notifyBiz.NotifyUserLikesOnComment(ctx, uid, recvUid, &bizsysnotify.NotifyLikesOnCommentReq{
				NoteId:    item.Oid,
				CommentId: commentId,
			})
			if err != nil {
				return xerror.Wrapf(err, "notify likes on comment").
					WithExtras("comment_id", commentId, "uid", uid, "recv", recvUid).WithCtx(ctx)
			}

			return nil
		},
	})
}
