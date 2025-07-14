package feed

import (
	"math/rand"
	"net/http"

	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/api-x/internal/handler/feed/biz"
	"github.com/ryanreadbooks/whimer/api-x/internal/handler/feed/model"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type Handler struct {
	bizz biz.FeedBiz
}

func NewHandler(c *config.Config) *Handler {
	return &Handler{
		bizz: biz.NewFeedBiz(),
	}
}

func (h *Handler) GetRecommend() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[model.FeedRecommendRequest](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, err := h.bizz.RandomFeed(r.Context(), req)
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
		req, err := xhttp.ParseValidate[model.FeedDetailRequest](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		noteId, err := req.NoteId.Uint64()
		if err != nil {
			xhttp.Error(r, w, xerror.ErrArgs.Msg("笔记不存在"))
			return
		}

		resp, err := h.bizz.GetNote(r.Context(), noteId)
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

		resp, page, err := h.bizz.ListNotesByUser(r.Context(), req.UserId, req.Cursor, req.Count)
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
