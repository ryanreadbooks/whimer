package backend

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/api-x/internal/backend/note"
	"github.com/ryanreadbooks/whimer/misc/errorx"

	notesdk "github.com/ryanreadbooks/whimer/note/sdk"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func (h *Handler) CreateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req note.CreateReq
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		// service to create note
		resp, err := note.GetNoter().CreateNote(r.Context(), req.AsPb())
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, note.CreateRes{NoteId: note.IdConfuser.ConfuseU(resp.NoteId)})
	}
}

func (h *Handler) UpdateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req note.UpdateReq
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		var noteId = note.IdConfuser.DeConfuseU(req.NoteId)

		_, err := note.GetNoter().UpdateNote(r.Context(), &notesdk.UpdateNoteReq{
			NoteId: noteId,
			Note: &notesdk.CreateNoteReq{
				Basic:  req.Basic.AsPb(),
				Images: req.Images.AsPb(),
			},
		})

		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, note.UpdateRes{NoteId: req.NoteId})
	}
}

func (h *Handler) DeleteNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req note.NoteIdReq
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		_, err := note.GetNoter().DeleteNote(r.Context(), &notesdk.DeleteNoteReq{
			NoteId: note.IdConfuser.DeConfuseU(req.NoteId),
		})

		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func (h *Handler) ListNotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := note.GetNoter().ListNote(r.Context(), &notesdk.ListNoteReq{})
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, note.NewListResFromPb(resp))
	}
}

func (h *Handler) GetNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req note.NoteIdReq
		if err := httpx.ParsePath(r, &req); err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		if len(req.NoteId) == 0 {
			httpx.Error(w, errorx.ErrArgs.Msg("笔记id错误"))
			return
		}

		resp, err := note.GetNoter().GetNote(r.Context(), &notesdk.GetNoteReq{
			NoteId: note.IdConfuser.DeConfuseU(req.NoteId),
		})
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, note.NewNoteItemFromPb(resp))
	}
}

func (h *Handler) UploadNoteAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req note.UploadAuthReq
		if err := httpx.ParseForm(r, &req); err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		resp, err := note.GetNoter().GetUploadAuth(r.Context(), req.AsPb())
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, resp)
	}
}
