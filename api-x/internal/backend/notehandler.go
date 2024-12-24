package backend

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/api-x/internal/backend/note"
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

		httpx.OkJson(w, note.CreateRes{NoteId: resp.NoteId})
	}
}

func (h *Handler) AdminUpdateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[note.UpdateReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		_, err = note.NoteCreatorServer().UpdateNote(r.Context(), &notev1.UpdateNoteRequest{
			NoteId: req.NoteId,
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
			NoteId: req.NoteId,
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
		req, err := xhttp.ParseValidate[note.ListReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}
		resp, err := note.NoteCreatorServer().ListNote(r.Context(), &notev1.ListNoteRequest{
			Cursor: req.Cursor,
			Count:  req.Count,
		})
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
			NoteId: req.NoteId,
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

		nid := req.NoteId
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

		nid := req.NoteId
		resp, err := note.NoteInteractServer().GetNoteLikes(r.Context(), &notev1.GetNoteLikesRequest{NoteId: nid})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, &note.GetLikesRes{
			Count:  resp.Likes,
			NoteId: resp.NoteId,
		})
	}
}
