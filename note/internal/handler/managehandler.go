package handler

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/svc"
	"github.com/ryanreadbooks/whimer/note/internal/types/manage"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func ManageCreateHandler(c *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req manage.CreateReq
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

		httpx.OkJson(w, manage.CreateRes{NoteId: noteId})
	}
}

func ManageUpdateHandler(c *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req manage.UpdateReq
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

		httpx.OkJson(w, manage.UpdateRes{NoteId: req.NoteId})
	}
}
