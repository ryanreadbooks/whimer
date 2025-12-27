package middleware

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/zeromicro/go-zero/rest"
)

func ApiOffline() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			xhttp.Error(r, w, xerror.ErrApiWentOffline)
		}
	}
}
