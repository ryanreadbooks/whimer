package note

import (
	"context"
	"net/http"

	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"

	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct{}

func NewHandler(c *config.Config) *Handler {
	return &Handler{}
}

func (h *Handler) hasNoteCheck(ctx context.Context, noteId uint64) error {
	if resp, err := NoteCreatorServer().IsNoteExist(ctx,
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

func (h *Handler) AdminCreateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[CreateReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		// service to create note
		resp, err := NoteCreatorServer().CreateNote(r.Context(), req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, CreateRes{NoteId: resp.NoteId})
	}
}

func (h *Handler) AdminUpdateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[UpdateReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		_, err = NoteCreatorServer().UpdateNote(r.Context(), &notev1.UpdateNoteRequest{
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

		httpx.OkJson(w, UpdateRes{NoteId: req.NoteId})
	}
}

func (h *Handler) AdminDeleteNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[NoteIdReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		_, err = NoteCreatorServer().DeleteNote(r.Context(), &notev1.DeleteNoteRequest{
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
		req, err := xhttp.ParseValidate[ListReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}
		resp, err := NoteCreatorServer().ListNote(r.Context(), &notev1.ListNoteRequest{
			Cursor: req.Cursor,
			Count:  req.Count,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, NewListResFromPb(resp))
	}
}

func (h *Handler) AdminGetNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[NoteIdReq](httpx.ParsePath, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		resp, err := NoteCreatorServer().GetNote(r.Context(), &notev1.GetNoteRequest{
			NoteId: req.NoteId,
		})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, NewAdminNoteItemFromPb(resp.Note))
	}
}

func (h *Handler) AdminUploadNoteAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[UploadAuthReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		resp, err := NoteCreatorServer().BatchGetUploadAuth(r.Context(), req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, resp)
	}
}

func (h *Handler) AdminUploadNoteAuthV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[UploadAuthReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		resp, err := NoteCreatorServer().BatchGetUploadAuthV2(r.Context(), req.AsPbV2())
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

		req, err := xhttp.ParseValidate[LikeReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		nid := req.NoteId
		_, err = NoteInteractServer().LikeNote(r.Context(), &notev1.LikeNoteRequest{
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
		req, err := xhttp.ParseValidate[NoteIdReq](httpx.ParsePath, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		nid := req.NoteId
		resp, err := NoteInteractServer().GetNoteLikes(r.Context(), &notev1.GetNoteLikesRequest{NoteId: nid})
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, &GetLikesRes{
			Count:  resp.Likes,
			NoteId: resp.NoteId,
		})
	}
}

// TODO 获取点赞过的笔记
func (h *Handler) GetLikeNotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
