package middleware

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/note/internal/external/passport"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	"github.com/ryanreadbooks/whimer/note/internal/svc"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func Auth(ctx *svc.ServiceContext) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			uid, _, err := passport.CheckSignIn(r.Context(), r)
			if err != nil {
				httpx.Error(w, err)
				return
			}

			newCtx := model.WithUid(r.Context(), uid)

			next(w, r.WithContext(newCtx))
		}
	}
}
