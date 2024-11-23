package http

import (
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

		resp, err := srv.Service.GetRecommendFeed(r.Context(), req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		httpx.OkJson(w, resp)
	}
}

func feedDetail() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

	}
}
