package feed

import (
	"math/rand"
	"net/http"

	"github.com/ryanreadbooks/whimer/api-x/internal/biz"
	bizfeed "github.com/ryanreadbooks/whimer/api-x/internal/biz/feed"
	feedmodel "github.com/ryanreadbooks/whimer/api-x/internal/biz/feed/model"
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/api-x/internal/model"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct {
	feedBiz *bizfeed.FeedBiz
}

func NewHandler(c *config.Config, bizz *biz.Biz) *Handler {
	return &Handler{
		feedBiz: bizz.FeedBiz,
	}
}

func (h *Handler) GetRecommend() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[feedmodel.FeedRecommendRequest](httpx.ParseForm, r)
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
		req, err := xhttp.ParseValidate[feedmodel.FeedDetailRequest](httpx.ParsePath, r)
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
		req, err := xhttp.ParseValidate[feedmodel.FeedByUserRequest](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, page, err := h.feedBiz.ListNotesByUser(r.Context(), req.UserId, int64(req.Cursor), req.Count)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		nextCursor := model.NoteId(page.NextCursor)
		if !page.HasNext {
			nextCursor = 0
		}

		xhttp.OkJson(w, &feedmodel.FeedByUserResponse{
			Items:      resp,
			NextCursor: nextCursor,
			HasNext:    page.HasNext,
		})
	}
}
