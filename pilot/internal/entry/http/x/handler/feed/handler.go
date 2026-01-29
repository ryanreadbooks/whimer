package feed

import (
	"math/rand"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/app"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notefeed"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notefeed/dto"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
	"github.com/ryanreadbooks/whimer/pilot/internal/domain/note/vo"

	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct {
	noteFeedApp *notefeed.Service
}

func NewHandler(c *config.Config, manager *app.Manager) *Handler {
	return &Handler{
		noteFeedApp: manager.NoteFeedApp,
	}
}

func (h *Handler) GetRecommend() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.GetRandomQuery](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, err := h.noteFeedApp.GetRandom(r.Context(), req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// make it random
		rand.Shuffle(len(resp), func(i, j int) { resp[i], resp[j] = resp[j], resp[i] })

		xhttp.OkJson(w, resp)
	}
}

func (h *Handler) GetNoteDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.GetFeedNoteQuery](httpx.ParsePath, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, err := h.noteFeedApp.GetFeedNote(r.Context(), req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

func (h *Handler) GetNotesByUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.ListUserFeedNotesQuery](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, page, err := h.noteFeedApp.ListUserFeedNotes(r.Context(), req.Uid, int64(req.Cursor), req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &dto.ListUserFeedNotesResult{
			Items:      resp,
			NextCursor: vo.NoteId(page.NextCursor),
			HasNext:    page.HasNext,
		})
	}
}

// 获取用户点赞过的笔记
func (h *Handler) GetLikedNotesByUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.ListUserLikedNoteQuery](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, page, err := h.noteFeedApp.ListUserLikedNotes(r.Context(), req.Uid, req.Cursor, req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &dto.ListUserLikedNoteResult{
			Items:      resp,
			NextCursor: page.NextCursor,
			HasNext:    page.HasNext,
		})
	}
}
