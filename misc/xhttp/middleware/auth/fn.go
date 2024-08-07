package auth

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/metadata"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func User(a *Auth) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			uid, sessId, err := a.User(r.Context(), r)
			if err != nil {
				httpx.Error(w, err)
				return
			}

			ctx := metadata.WithUid(r.Context(), uid)
			ctx = metadata.WithSessId(ctx, sessId)

			next(w, r.WithContext(ctx))
		}
	}
}

func UserWeb(a *Auth) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			uid, sessId, err := a.UserWeb(r.Context(), r)
			if err != nil {
				httpx.Error(w, err)
				return
			}

			ctx := metadata.WithUid(r.Context(), uid)
			ctx = metadata.WithSessId(ctx, sessId)

			next(w, r.WithContext(ctx))
		}
	}
}
