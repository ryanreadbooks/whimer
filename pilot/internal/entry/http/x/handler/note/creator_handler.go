package note

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	commondto "github.com/ryanreadbooks/whimer/pilot/internal/app/common/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notecreator/dto"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func (h *Handler) CreatorCreateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		command, err := xhttp.ParseValidateJsonBody[dto.CreateNoteCommand](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := h.noteCreatorApp.CreateNote(ctx, command)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

func (h *Handler) CreatorUpdateNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		command, err := xhttp.ParseValidateJsonBody[dto.UpdateNoteCommand](r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		err = h.noteCreatorApp.UpdateNote(ctx, command)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, dto.UpdateNoteResult{NoteId: command.NoteId})
	}
}

// 删除笔记
func (h *Handler) CreatorDeleteNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[commondto.NoteIdReq](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()

		err = h.noteCreatorApp.DeleteNote(ctx, req.NoteId)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

// 分页列出个人笔记
func (h *Handler) CreatorPageListNotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query, err := xhttp.ParseValidate[dto.PageListNotesQuery](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		result, err := h.noteCreatorApp.PageListNotes(ctx, query)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, result)
	}
}

// 获取个人笔记
func (h *Handler) CreatorGetNote() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[commondto.NoteIdReq](httpx.ParsePath, r)
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg(err.Error()))
			return
		}

		ctx := r.Context()
		result, err := h.noteCreatorApp.GetNote(ctx, req.NoteId)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, result)
	}
}

// 创建新标签
func (h *Handler) AddNewTag() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.AddTagCommand](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		ctx := r.Context()
		resp, err := h.noteCreatorApp.AddTag(ctx, req.Name)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

// 搜索笔记标签
func (h *Handler) SearchTags() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.SearchTagsQuery](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		if req.Name == "" {
			xhttp.OkJson(w, []*dto.SearchedTag{})
			return
		}

		tags, err := h.noteCreatorApp.SearchTags(r.Context(), req.Name)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, tags)
	}
}
