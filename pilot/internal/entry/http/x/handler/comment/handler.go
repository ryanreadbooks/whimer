package comment

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	bizcomment "github.com/ryanreadbooks/whimer/pilot/internal/biz/comment"
	bizstorage "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/storage"
	bizuser "github.com/ryanreadbooks/whimer/pilot/internal/biz/common/user"
	bizsearch "github.com/ryanreadbooks/whimer/pilot/internal/biz/search"
	bizsysnotify "github.com/ryanreadbooks/whimer/pilot/internal/biz/sysnotify"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
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
		req, err := xhttp.ParseValidate[PubReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		commentId, err := h.commentBiz.PublishNoteComment(ctx, req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		h.afterNoteCommented(ctx, commentId, req)
		httpx.OkJson(w, &PubRes{CommentId: commentId})
	}
}

// 只获取主评论
func (h *Handler) PageGetNoteRootComments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[GetCommentsReq](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		if err := h.checkHasNote(ctx, int64(req.Oid)); err != nil {
			xhttp.Error(r, w, err)
			return
		}

		res, err := h.commentBiz.PageGetNoteRootComments(ctx, req.AsPb())
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
		req, err := xhttp.ParseValidate[GetSubCommentsReq](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		if err := h.checkHasNote(ctx, int64(req.Oid)); err != nil {
			xhttp.Error(r, w, err)
			return
		}

		res, err := h.commentBiz.PageGetNoteSubComments(ctx, req.AsPb())
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
		req, err := xhttp.ParseValidate[GetCommentsReq](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		if err := h.checkHasNote(r.Context(), int64(req.Oid)); err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		res, err := h.commentBiz.PageGetNoteComments(ctx, &bizcomment.PageGetNoteCommentsReq{
			Oid:    int64(req.Oid),
			Cursor: req.Cursor,
			SortBy: int32(req.SortBy),
			SeekId: req.SeekId,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, res)
	}
}

func (h *Handler) DelNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[DelReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		err = h.commentBiz.DelNoteComment(ctx, req.CommentId, int64(req.Oid))
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
		req, err := xhttp.ParseValidate[PinReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		err = h.commentBiz.PinNoteComment(ctx, int64(req.Oid), req.CommentId, int8(req.Action))
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) LikeNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[ThumbUpReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		err = h.commentBiz.LikeNoteComment(ctx, req.CommentId, uint8(req.Action))
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		if req.Action == ThumbActionDo {
			// 通知用户
			h.asyncNotifyLikeComment(ctx, req.CommentId)
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) DislikeNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[ThumbDownReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		err = h.commentBiz.DislikeNoteComment(ctx, req.CommentId, uint8(req.Action))
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) GetNoteCommentLikeCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[GetLikeCountReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		count, err := h.commentBiz.GetNoteCommentLikeCount(ctx, req.CommentId)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, &GetLikeCountRes{Comment: req.CommentId, Likes: count})
	}
}

func (h *Handler) UploadCommentImages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[UploadImagesReq](httpx.ParseForm, r)
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
