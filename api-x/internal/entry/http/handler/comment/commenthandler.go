package comment

import (
	"context"
	"net/http"

	"github.com/ryanreadbooks/whimer/api-x/internal/biz"
	bizcomment "github.com/ryanreadbooks/whimer/api-x/internal/biz/comment"
	bizmodel "github.com/ryanreadbooks/whimer/api-x/internal/biz/comment/model"
	bizsearch "github.com/ryanreadbooks/whimer/api-x/internal/biz/search"
	bizuser "github.com/ryanreadbooks/whimer/api-x/internal/biz/user"
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct {
	userBiz    *bizuser.Biz
	searchBiz  *bizsearch.Biz
	commentBiz *bizcomment.Biz
}

func NewHandler(c *config.Config, bizz *biz.Biz) *Handler {
	return &Handler{
		userBiz:    bizz.UserBiz,
		searchBiz:  bizz.SearchBiz,
		commentBiz: bizz.CommentBiz,
	}
}

func (h *Handler) syncCommentCountToSearcher(ctx context.Context, noteId string, incr int64) {
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "note.handler.commentnote.synces",
		Job: func(ctx context.Context) error {
			err := h.searchBiz.NoteStatSyncer.AddCommentCount(ctx, noteId, incr)
			if err != nil {
				xlog.Msg("note stat add comment count failed").
					Extras("note_id", noteId, "incr", incr).
					Err(err).Errorx(ctx)
			}

			return err
		},
	})
}

// 发表评论
func (h *Handler) PublishNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[bizmodel.PubReq](httpx.ParseJsonBody, r)
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

		httpx.OkJson(w, res)
	}
}

// 只获取主评论
func (h *Handler) PageGetNoteRootComments() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[bizmodel.GetCommentsReq](httpx.Parse, r)
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
		req, err := xhttp.ParseValidate[bizmodel.GetSubCommentsReq](httpx.Parse, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		if err := h.checkHasNote(r.Context(), int64(req.Oid)); err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
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
		req, err := xhttp.ParseValidate[bizmodel.GetCommentsReq](httpx.Parse, r)
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
		req, err := xhttp.ParseValidate[bizmodel.DelReq](httpx.ParseJsonBody, r)
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
		req, err := xhttp.ParseValidate[bizmodel.PinReq](httpx.ParseJsonBody, r)
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
		req, err := xhttp.ParseValidate[bizmodel.ThumbUpReq](httpx.ParseJsonBody, r)
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

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) DislikeNoteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[bizmodel.ThumbDownReq](httpx.ParseJsonBody, r)
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
		req, err := xhttp.ParseValidate[bizmodel.GetLikeCountReq](httpx.ParseForm, r)
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
	if resp, err := infra.NoteCreatorServer().IsNoteExist(ctx,
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
		req, err := xhttp.ParseValidate[bizmodel.UploadImagesReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := h.commentBiz.UploadCommentImages(ctx, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

func (h *Handler) MentionUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[MentionUserReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		var (
			ctx = r.Context()
			uid = metadata.Uid(ctx)
		)

		res, err := h.userBiz.BrutalListFollowingsByName(ctx, uid, req.Search)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, res)
	}
}
