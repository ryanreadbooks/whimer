package middleware

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/api-x/internal/infra"
	"github.com/ryanreadbooks/whimer/misc/metadata"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xhttp"
	"github.com/ryanreadbooks/whimer/passport/pkg/middleware/auth"
	"github.com/zeromicro/go-zero/rest"
)

// 必须登录
func MustLogin() rest.Middleware {
	return auth.UserWeb(infra.Auther())
}

// 可以不用登录 也可以登录
func CanLogin() rest.Middleware {
	return auth.UserWebOptional(infra.Auther())
}

func MustLoginCheck() rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 仅检查
			ctx := r.Context()
			ok := metadata.HasUid(ctx) &&
				metadata.HasSessId(ctx) &&
				metadata.HasUserNickname(ctx)

			if !ok {
				xhttp.Error(r, w, xerror.ErrNotLogin)
				return
			}

			next(w, r)
		}
	}
}
