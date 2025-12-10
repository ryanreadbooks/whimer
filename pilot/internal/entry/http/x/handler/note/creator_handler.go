package note

import (
	"context"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"
	notemodel "github.com/ryanreadbooks/whimer/pilot/internal/biz/note/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/model"
	"github.com/ryanreadbooks/whimer/pilot/internal/model/uploadresource"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func (h *Handler) handleAssetResource(ctx context.Context, req *CreateReq) (err error, rollback func()) {
	rollback = func() {}

	// check every resource
	switch req.Basic.Type {
	case model.NoteTypeImage:
		images := make([]notemodel.NoteImage, 0, len(req.Images))
		for _, img := range req.Images {
			images = append(images, notemodel.NoteImage{
				FileId: img.FileId,
				Width:  img.Width,
				Height: img.Height,
				Format: img.Format,
			})
		}
		err = h.noteBiz.MarkNoteImages(ctx, images)
		rollback = func() {
			concurrent.SimpleSafeGo(
				ctx,
				"note_create_fail_unmark_resources",
				func(ctx context.Context) error {
					return h.noteBiz.UnmarkNoteImages(ctx, images)
				})
		}
	case model.NoteTypeVideo:
		err = h.noteBiz.CheckNoteVideo(ctx)
	}

	return
}

// 发布笔记
func (h *Handler) CreatorCreateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[CreateReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}
		ctx := r.Context()

		// check every resource
		err, rollback := h.handleAssetResource(ctx, req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, err := h.noteBiz.CreateNote(ctx, req.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			rollback()
			return
		}

		note, err := h.noteBiz.GetNote(ctx, int64(resp.NoteId))
		if err == nil && note != nil {
			h.noteBiz.AfterNoteUpserted(ctx, note)
		}

		xhttp.OkJson(w, CreateRes{NoteId: resp.NoteId})
	}
}

// 修改笔记
func (h *Handler) CreatorUpdateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[UpdateReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		// check new resources
		err, rollback := h.handleAssetResource(ctx, &req.CreateReq)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		err = h.noteBiz.UpdateNote(ctx, req.NoteId, req.CreateReq.AsPb())
		if err != nil {
			xhttp.Error(r, w, err)
			rollback()
			return
		}

		note, err := h.noteBiz.GetNote(ctx, int64(req.NoteId))
		if err == nil && note != nil {
			h.noteBiz.AfterNoteUpserted(ctx, note)
		}

		xhttp.OkJson(w, UpdateRes{NoteId: req.NoteId})
	}
}

// 删除笔记
func (h *Handler) CreatorDeleteNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[NoteIdReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		// get first
		note, err := h.noteBiz.GetNote(ctx, int64(req.NoteId))
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		err = h.noteBiz.DeleteNote(ctx, req.NoteId)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		concurrent.SimpleSafeGo(ctx,
			"note_delete_fail_unmark_resources",
			func(ctx context.Context) error {
				switch note.NoteType {
				case notev1.NoteAssetType_IMAGE:
					return h.noteBiz.UnmarkNoteImages(ctx, notemodel.NoteImagesFromPbs(note.Images))
				case notev1.NoteAssetType_VIDEO:
					// todo
				}
				return nil
			})

		h.noteBiz.AsyncDeleteNoteFromSearcher(ctx, int64(req.NoteId))

		httpx.OkJson(w, nil)
	}
}

// 分页列出个人笔记
func (h *Handler) CreatorPageListNotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[PageListReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := h.noteBiz.PageListNotes(ctx, req.Page, req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		result := NewPageListResFromPb(resp)
		h.noteBiz.AssignNoteExtra(ctx, result.Items)

		xhttp.OkJson(w, result)
	}
}

// 获取个人笔记
func (h *Handler) CreatorGetNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[NoteIdReq](httpx.ParsePath, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		note, err := h.noteBiz.GetNote(ctx, int64(req.NoteId))
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		result := model.NewAdminNoteItemFromPb(note)
		h.noteBiz.AssignNoteExtra(ctx, []*model.AdminNoteItem{result})
		xhttp.OkJson(w, result)
	}
}

// Deprecated
//
// See: upload.GetTempCreds for newest usage
func (h *Handler) CreatorUploadNoteAuthV2() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[UploadAuthReq](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		resp, err := h.storageBiz.RequestUploadTicket(ctx, uploadresource.NoteImage, req.Count, req.Source)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}
