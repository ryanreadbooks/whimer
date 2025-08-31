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
			user, sessId, err := a.User(r.Context(), r)
			if err != nil {
				httpx.Error(w, err)
				return
			}

			ctx := metadata.WithUid(r.Context(), user.Uid)
			ctx = metadata.WithSessId(ctx, sessId)
			ctx = metadata.WithUserNickname(ctx, user.Nickname)

			next(w, r.WithContext(ctx))
		}
	}
}

// http接口认证中间件 并且注入uid
func UserWeb(a *Auth) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			user, sessId, err := a.UserWeb(ctx, r)
			if err != nil {
				httpx.Error(w, err)
				return
			}

			ctx = metadata.WithUid(ctx, user.Uid)
			ctx = metadata.WithSessId(ctx, sessId)
			ctx = metadata.WithUserNickname(ctx, user.Nickname)

			next(w, r.WithContext(ctx))
		}
	}
}

// 可选登录接口
func UserWebOptional(a *Auth) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			user, sessId, err := a.UserWeb(ctx, r)
			// err is allowed
			if err == nil && user != nil {
				ctx = metadata.WithUid(ctx, user.Uid)
				ctx = metadata.WithSessId(ctx, sessId)
				ctx = metadata.WithUserNickname(ctx, user.Nickname)
			}

			next(w, r.WithContext(ctx))
		}
	}
}
