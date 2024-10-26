package backend

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/api-x/internal/backend/note"
	"github.com/ryanreadbooks/whimer/misc/errorx"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xhttp"

	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func (h *Handler) CreateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[note.CreateReq](httpx.ParseJsonBody, r)
		if err != nil {
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
		req, err := xhttp.ParseValidate[note.UpdateReq](httpx.ParseJsonBody, r)
		if err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		var noteId = note.IdConfuser.DeConfuseU(req.NoteId)

		_, err = note.GetNoter().UpdateNote(r.Context(), &notev1.UpdateNoteReq{
			NoteId: noteId,
			Note: &notev1.CreateNoteReq{
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
		req, err := xhttp.ParseValidate[note.NoteIdReq](httpx.ParseJsonBody, r)
		if err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		_, err = note.GetNoter().DeleteNote(r.Context(), &notev1.DeleteNoteReq{
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
		resp, err := note.GetNoter().ListNote(r.Context(), &notev1.ListNoteReq{})
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, note.NewListResFromPb(resp))
	}
}

func (h *Handler) GetNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[note.NoteIdReq](httpx.ParsePath, r)
		if err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		resp, err := note.GetNoter().GetNote(r.Context(), &notev1.GetNoteReq{
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
		req, err := xhttp.ParseValidate[note.UploadAuthReq](httpx.ParseForm, r)
		if err != nil {
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

// 点赞/取消点赞笔记
func (h *Handler) LikeNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			uid = metadata.Uid(r.Context())
		)

		req, err := xhttp.ParseValidate[note.LikeReq](httpx.ParseJsonBody, r)
		if err != nil {
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		nid := note.IdConfuser.DeConfuseU(req.NoteId)
		_, err = note.GetNoter().LikeNote(r.Context(), &notev1.LikeNoteReq{
			NoteId:    nid,
			Uid:       uid,
			Operation: notev1.LikeNoteReq_Operation(req.Action),
		})
		if err != nil {
			httpx.Error(w, err)
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
			httpx.Error(w, errorx.ErrArgs.Msg(err.Error()))
			return
		}

		nid := note.IdConfuser.DeConfuseU(req.NoteId)
		resp, err := note.GetNoter().GetNoteLikes(r.Context(), &notev1.GetNoteLikesReq{NoteId: nid})
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, &note.GetLikesRes{
			Count:  resp.Likes,
			NoteId: note.IdConfuser.ConfuseU(resp.NoteId),
		})
	}
}
