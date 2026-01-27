package comment

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/app"
	appcomment "github.com/ryanreadbooks/whimer/pilot/internal/app/comment"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/comment/dto"
	appuser "github.com/ryanreadbooks/whimer/pilot/internal/app/user"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	storagevo "github.com/ryanreadbooks/whimer/pilot/internal/domain/common/storage/vo"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct {
	userApp    *appuser.Service
	commentApp *appcomment.Service
}

func NewHandler(c *config.Config, manager *app.Manager) *Handler {
	return &Handler{
		userApp:    manager.UserApp,
		commentApp: manager.CommentApp,
	}
}

// 发表评论
func (h *Handler) PublishNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd, err := xhttp.ParseValidate[dto.PublishCommentCommand](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		commentId, err := h.commentApp.PublishComment(ctx, cmd)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, &dto.PublishCommentResult{CommentId: commentId})
	}
}

// 只获取主评论
func (h *Handler) PageGetNoteRootComments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q, err := xhttp.ParseValidate[dto.GetCommentsQuery](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()

		res, err := h.commentApp.PageGetRootComments(ctx, q)
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
		q, err := xhttp.ParseValidate[dto.GetSubCommentsQuery](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		res, err := h.commentApp.PageGetSubComments(ctx, q)
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
		q, err := xhttp.ParseValidate[dto.GetCommentsQuery](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		res, err := h.commentApp.PageGetComments(ctx, q)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, res)
	}
}

func (h *Handler) DelNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd, err := xhttp.ParseValidate[dto.DeleteCommentCommand](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		err = h.commentApp.DeleteComment(ctx, cmd)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) PinNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd, err := xhttp.ParseValidate[dto.PinCommentCommand](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		err = h.commentApp.PinComment(ctx, cmd)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) LikeNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd, err := xhttp.ParseValidate[dto.LikeCommentCommand](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		err = h.commentApp.LikeComment(ctx, cmd)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) DislikeNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd, err := xhttp.ParseValidate[dto.DislikeCommentCommand](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		err = h.commentApp.DislikeComment(ctx, cmd)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) GetNoteCommentLikeCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q, err := xhttp.ParseValidate[dto.GetLikeCountQuery](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		res, err := h.commentApp.GetCommentLikeCount(ctx, q.CommentId)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, &dto.GetLikeCountResult{CommentId: res.CommentId, Likes: res.Likes})
	}
}

func (h *Handler) UploadCommentImages() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cmd, err := xhttp.ParseValidate[dto.UploadImagesCommand](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := h.commentApp.GetUploadImageTicket(ctx, cmd.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, newUploadTicket(resp))
	}
}

func newUploadTicket(t *storagevo.UploadTicketDeprecated) *dto.UploadTicket {
	return &dto.UploadTicket{
		StoreKeys:   t.FileIds,
		CurrentTime: t.CurrentTime,
		ExpireTime:  t.ExpireTime,
		UploadAddr:  t.UploadAddr,
		Token:       t.Token,
	}
}
