package backend

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/ryanreadbooks/whimer/api-x/internal/backend/comment"
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/note"
	commentv1 "github.com/ryanreadbooks/whimer/comment/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"

	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func (h *Handler) AdminCreateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[note.CreateReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		// service to create note
		resp, err := note.NoteCreatorServer().CreateNote(r.Context(), req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, note.CreateRes{NoteId: note.IdConfuser.ConfuseU(resp.NoteId)})
	}
}

func (h *Handler) AdminUpdateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[note.UpdateReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		var noteId = note.IdConfuser.DeConfuseU(req.NoteId)

		_, err = note.NoteCreatorServer().UpdateNote(r.Context(), &notev1.UpdateNoteRequest{
			NoteId: noteId,
			Note: &notev1.CreateNoteRequest{
				Basic:  req.Basic.AsPb(),
				Images: req.Images.AsPb(),
			},
		})

		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, note.UpdateRes{NoteId: req.NoteId})
	}
}

func (h *Handler) AdminDeleteNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[note.NoteIdReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		_, err = note.NoteCreatorServer().DeleteNote(r.Context(), &notev1.DeleteNoteRequest{
			NoteId: note.IdConfuser.DeConfuseU(req.NoteId),
		})

		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) AdminListNotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := note.NoteCreatorServer().ListNote(r.Context(), &notev1.ListNoteRequest{})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, note.NewListResFromPb(resp))
	}
}

func (h *Handler) AdminGetNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[note.NoteIdReq](httpx.ParsePath, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		resp, err := note.NoteCreatorServer().GetNote(r.Context(), &notev1.GetNoteRequest{
			NoteId: note.IdConfuser.DeConfuseU(req.NoteId),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, note.NewAdminNoteItemFromPb(resp.Note))
	}
}

func (h *Handler) AdminUploadNoteAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[note.UploadAuthReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		resp, err := note.NoteCreatorServer().GetUploadAuth(r.Context(), req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, resp)
	}
}

// 点赞/取消点赞笔记
func (h *Handler) LikeNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			uid = metadata.Uid(r.Context())
		)

		req, err := xhttp.ParseValidate[note.LikeReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		nid := note.IdConfuser.DeConfuseU(req.NoteId)
		_, err = note.NoteInteractServer().LikeNote(r.Context(), &notev1.LikeNoteRequest{
			NoteId:    nid,
			Uid:       uid,
			Operation: notev1.LikeNoteRequest_Operation(req.Action),
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}
		httpx.OkJson(w, nil)
	}
}

// 获取笔记点赞数量
func (h *Handler) GetNoteLikeCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[note.NoteIdReq](httpx.ParsePath, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		nid := note.IdConfuser.DeConfuseU(req.NoteId)
		resp, err := note.NoteInteractServer().GetNoteLikes(r.Context(), &notev1.GetNoteLikesRequest{NoteId: nid})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, &note.GetLikesRes{
			Count:  resp.Likes,
			NoteId: note.IdConfuser.ConfuseU(resp.NoteId),
		})
	}
}

// 获取笔记
func (h *Handler) GetNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[note.NoteIdReq](httpx.ParsePath, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		var (
			uid    = metadata.Uid(r.Context())
			noteId = note.IdConfuser.DeConfuseU(req.NoteId)

			wg    sync.WaitGroup
			resp1 *commentv1.CountReplyRes
			resp2 *commentv1.CheckUserCommentOnObjectResponse
		)
		wg.Add(2)

		// 获取评论数
		concurrent.DoneInCtx(r.Context(), time.Second*10, func(ctx context.Context) {
			defer wg.Done()
			resp1, _ = comment.Commenter().CountReply(ctx, &commentv1.CountReplyReq{Oid: noteId})
		})

		concurrent.DoneInCtx(r.Context(), time.Second*10, func(ctx context.Context) {
			defer wg.Done()
			resp2, _ = comment.Commenter().CheckUserCommentOnObject(ctx, &commentv1.CheckUserCommentOnObjectRequest{
				Uid: uid,
				Oid: noteId,
			})
		})

		resp, err := note.NoteFeedServer().GetFeedNote(r.Context(), &notev1.GetFeedNoteRequest{
			NoteId: noteId,
		})

		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		wg.Wait()

		feed := note.NewFeedNoteItemFromPb(resp.Item)
		// 注入评论笔记的评论信息
		feed.Comments = resp1.GetNumReply()
		feed.Interact.Commented = resp2.GetCommented()

		httpx.OkJson(w, feed)
	}
}
