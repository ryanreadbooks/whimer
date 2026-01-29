package feed

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/pilot/internal/app/notefeed/dto"

	"github.com/zeromicro/go-zero/rest/httpx"
)

// 搜索笔记
func (h *Handler) SearchNotes() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := xhttp.ParseValidate[dto.SearchNotesQuery](httpx.ParseJsonBody, r)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		resp, err := h.noteFeedApp.SearchNotes(r.Context(), req)
		if err != nil {
			xhttp.Error(r, w, err)
			return
		}

		xhttp.OkJson(w, resp)
	}
}

// 获取搜索可用的过滤器
func (h *Handler) GetSearchNotesAvailableFilters() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
	}
}
