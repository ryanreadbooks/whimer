package feed

import (
	"math/rand"
	"net/http"

	bizfeed "github.com/ryanreadbooks/whimer/api-x/internal/biz/feed"
	"github.com/ryanreadbooks/whimer/api-x/internal/biz/feed/model"
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct {
	feedBiz bizfeed.FeedBiz
}

func NewHandler(c *config.Config, feedBiz bizfeed.FeedBiz) *Handler {
	return &Handler{
		feedBiz: feedBiz,
	}
}

func (h *Handler) GetRecommend() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[model.FeedRecommendRequest](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, err := h.feedBiz.RandomFeed(r.Context(), req)
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
		req, err := xhttp.ParseValidate[model.FeedDetailRequest](httpx.ParsePath, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, err := h.feedBiz.GetNote(r.Context(), int64(req.NoteId))
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

func (h *Handler) GetNotesByUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[model.FeedByUserRequest](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, page, err := h.feedBiz.ListNotesByUser(r.Context(), req.UserId, req.Cursor, req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &model.FeedByUserResponse{
			Items:      resp,
			NextCursor: page.NextCursor,
			HasNext:    page.HasNext,
		})
	}
}

// 获取点赞过的笔记
func (h *Handler) GetLikedNotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			ctx = r.Context()
		)

		req, err := xhttp.ParseValidate[model.GetLikedNoteRequest](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, page, err := h.feedBiz.ListLikedNotes(ctx, req.Cursor, req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, &model.GetLikedNoteResponse{
			Items:      resp,
			NextCursor: page.NextCursor,
			HasNext:    page.HasNext,
		})
	}
}
