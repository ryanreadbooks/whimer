package http

import (
	"math/rand"
	"net/http"

	"github.com/ryanreadbooks/whimer/feed/internal/model"
	"github.com/ryanreadbooks/whimer/feed/internal/srv"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func feedRecommend() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[model.FeedRecommendRequest](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, err := srv.Service.FeedBiz.RandomFeed(r.Context(), req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		// make it random
		rand.Shuffle(len(resp), func(i, j int) { resp[i], resp[j] = resp[j], resp[i] })

		httpx.OkJson(w, resp)
	}
}

func feedDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[model.FeedDetailRequest](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, err := srv.Service.FeedBiz.GetNote(r.Context(), req.NoteId)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, resp)
	}
}

func feedByUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[model.FeedByUserRequest](httpx.ParseForm, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, page, err := srv.Service.FeedBiz.ListNotesByUser(r.Context(), req.UserId, req.Cursor, req.Count)
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
