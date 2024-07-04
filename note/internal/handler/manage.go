package handler

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/svc"
	mgtp "github.com/ryanreadbooks/whimer/note/internal/types/manage"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// /note/v1/manage/create
func ManageCreateHandler(c *svc.ServiceContext) http.HandlerFunc {
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
		// TODO get uid from wherever
		noteId, err := c.Manage.Create(r.Context(), 100, &req)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, mgtp.CreateRes{NoteId: noteId})
	}
}

// /note/v1/manage/update
func ManageUpdateHandler(c *svc.ServiceContext) http.HandlerFunc {
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

		// TODO get uid from whatever
		err := c.Manage.Update(r.Context(), 100, &req)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, mgtp.UpdateRes{NoteId: req.NoteId})
	}
}

func ManageDeleteHandler(c *svc.ServiceContext) http.HandlerFunc {
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

		err := c.Manage.Delete(r.Context(), 100, &req)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, nil)
	}
}

func ManageListHandler(c *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := c.Manage.List(r.Context(), 100)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, res)
	}
}

func ManageGetNoteHandler(c *svc.ServiceContext) http.HandlerFunc {
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

		res, err := c.Manage.GetNote(r.Context(), 100, req.NoteId)
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

		res, err := c.Manage.UploadAuth(r.Context(), &req)
		if err != nil {
			httpx.Error(w, err)
			return
		}

		httpx.OkJson(w, res)
	}
}
