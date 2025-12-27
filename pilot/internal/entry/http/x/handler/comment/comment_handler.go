package comment

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	bizcomment "github.com/ryanreadbooks/whimer/pilot/internal/biz/comment"
	commentmodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/comment/model"
	bizsearch "github.com/ryanreadbooks/whimer/pilot/internal/biz/search"
	bizstorage "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/storage"
	bizsysnotify "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify"
	sysnotifymodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify/model"
	bizuser "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/user"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/pilot/internal/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/uploadresource"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct {
	userBiz    *bizuser.Biz
	searchBiz  *bizsearch.Biz
	commentBiz *bizcomment.Biz
	notifyBiz  *bizsysnotify.Biz
	storageBiz *bizstorage.Biz
}

func NewHandler(c *config.Config, bizz *biz.Biz) *Handler {
	return &Handler{
		userBiz:    bizz.UserBiz,
		searchBiz:  bizz.SearchBiz,
		commentBiz: bizz.CommentBiz,
		notifyBiz:  bizz.SysNotifyBiz,
		storageBiz: bizz.UploadBiz,
	}
}

// 发表评论
func (h *Handler) PublishNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[commentmodel.PubReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		res, err := h.commentBiz.PublishNoteComment(ctx, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		noteId := req.Oid.String()
		h.syncCommentCountToSearcher(ctx, noteId, 1)
		// 通知被@的用户
		h.notifyWhenAtUsers(ctx, res.CommentId, req)
		h.asyncNotifyReplyUser(ctx, res.CommentId, req)
		h.appendRecentContacts(ctx, req.AtUsers)

		httpx.OkJson(w, res)
	}
}

// 只获取主评论
func (h *Handler) PageGetNoteRootComments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[commentmodel.GetCommentsReq](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		if err := h.checkHasNote(ctx, int64(req.Oid)); err != nil {
			xhttp.Error(r, w, err)
			return
		}

		res, err := h.commentBiz.PageGetNoteRootComments(ctx, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, res)
	}
}

// 只获取子评论
func (h *Handler) PageGetNoteSubComments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[commentmodel.GetSubCommentsReq](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		if err := h.checkHasNote(ctx, int64(req.Oid)); err != nil {
			xhttp.Error(r, w, err)
			return
		}

		res, err := h.commentBiz.PageGetNoteSubComments(ctx, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, res)
	}
}

// 获取主评论信息（包含其下子评论）
func (h *Handler) PageGetNoteComments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[commentmodel.GetCommentsReq](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		if err := h.checkHasNote(r.Context(), int64(req.Oid)); err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		res, err := h.commentBiz.PageGetNoteComments(ctx, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, res)
	}
}

func (h *Handler) DelNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[commentmodel.DelReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		err = h.commentBiz.DelNoteComment(ctx, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		noteId := req.Oid.String()
		h.syncCommentCountToSearcher(ctx, noteId, -1)

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) PinNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[commentmodel.PinReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		err = h.commentBiz.PinNoteComment(ctx, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) LikeNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[commentmodel.ThumbUpReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		err = h.commentBiz.LikeNoteComment(ctx, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		if req.Action == commentmodel.ThumbActionDo {
			// 通知用户
			h.asyncNotifyLikeComment(ctx, req.CommentId)
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) DislikeNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[commentmodel.ThumbDownReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		err = h.commentBiz.DislikeNoteComment(ctx, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) GetNoteCommentLikeCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[commentmodel.GetLikeCountReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		res, err := h.commentBiz.GetNoteCommentLikeCount(ctx, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, res)
	}
}

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

func (h *Handler) UploadCommentImages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[commentmodel.UploadImagesReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := h.storageBiz.RequestUploadTicket(ctx, uploadresource.CommentImage, req.Count, "")
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, newUploadTicket(resp))
	}
}

// 异步写入最近联系人
func (h *Handler) appendRecentContacts(ctx context.Context, atUsers model.AtUserList) {
	var (
		uid = metadata.Uid(ctx)
	)

	h.userBiz.AsyncAppendRecentContactsAtUser(ctx, uid, atUsers)
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

func (h *Handler) notifyWhenAtUsers(ctx context.Context, commentId int64, req *commentmodel.PubReq) {
	var (
		uid = metadata.Uid(ctx)
	)

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

func (h *Handler) asyncNotifyReplyUser(ctx context.Context, triggerComment int64, req *commentmodel.PubReq) {
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
				xlog.Msg("notify reply to user failed whtn get comment content").
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
	var uid = metadata.Uid(ctx)

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
