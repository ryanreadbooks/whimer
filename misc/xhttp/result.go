package xhttp

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func Error(r *http.Request, w http.ResponseWriter, e error) {
	httpx.ErrorCtx(r.Context(), w, e)
}

func OkJson(w http.ResponseWriter, data any) {
	httpx.OkJson(w, data)
}
