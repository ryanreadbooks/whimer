package http

import (
	"net/http"
	stdhttp "net/http"

	"github.com/ryanreadbooks/whimer/feed/internal/srv"
)

func feedRecommend(s *srv.Service) http.HandlerFunc {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		
	}
}

func feedDetail(s *srv.Service) http.HandlerFunc {
	return func(w stdhttp.ResponseWriter, r *stdhttp.Request) {

	}
}
