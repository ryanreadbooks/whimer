package handler

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/note/internal/global"
	mgtp "github.com/ryanreadbooks/whimer/note/internal/model/creator"
	"github.com/ryanreadbooks/whimer/note/internal/svc"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func CreatorCreateHandler(c *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req mgtp.CreateReq
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.Error(w, global.ErrArgs.Msg(err.Error()))
			return
		}

		if err := req.Validate(); err != nil {
			httpx.Error(w, err)
			return
		}

		// service to create note
		noteId, err := c.CreatorSvc.Create(r.Context(), &req)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, mgtp.CreateRes{NoteId: noteId})
	}
}

func CreatorUpdateHandler(c *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req mgtp.UpdateReq
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.Error(w, global.ErrArgs.Msg(err.Error()))
			return
		}

		if err := req.Validate(); err != nil {
			httpx.Error(w, err)
			return
		}

		err := c.CreatorSvc.Update(r.Context(), &req)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, mgtp.UpdateRes{NoteId: req.NoteId})
	}
}

func CreatorDeleteHandler(c *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req mgtp.DeleteReq
		if err := httpx.ParseJsonBody(r, &req); err != nil {
			httpx.Error(w, global.ErrArgs.Msg(err.Error()))
			return
		}

		if err := req.Validate(); err != nil {
			httpx.Error(w, err)
			return
		}

		err := c.CreatorSvc.Delete(r.Context(), &req)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func CreatorListHandler(c *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := c.CreatorSvc.List(r.Context())
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, res)
	}
}

func CreatorGetNoteHandler(c *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req mgtp.GetNoteReq
		if err := httpx.ParsePath(r, &req); err != nil {
			httpx.Error(w, global.ErrArgs.Msg(err.Error()))
			return
		}

		if req.NoteId == "" {
			httpx.Error(w, global.ErrArgs.Msg("笔记不存在"))
			return
		}

		res, err := c.CreatorSvc.GetNote(r.Context(), req.NoteId)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, res)
	}
}

// /note/v1/upload/auth
func UploadAuthHandler(c *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req mgtp.UploadAuthReq
		if err := httpx.ParseForm(r, &req); err != nil {
			httpx.Error(w, global.ErrArgs.Msg(err.Error()))
			return
		}

		if err := req.Validate(); err != nil {
			httpx.Error(w, err)
			return
		}

		res, err := c.CreatorSvc.UploadAuth(r.Context(), &req)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, res)
	}
}
