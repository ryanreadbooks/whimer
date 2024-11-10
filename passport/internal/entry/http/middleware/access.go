package middleware

import (
	"net/http"
	"net/url"

	global "github.com/ryanreadbooks/whimer/passport/internal/global"
	"github.com/ryanreadbooks/whimer/passport/internal/model"
	"github.com/ryanreadbooks/whimer/passport/internal/srv"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// 登录验证中间件
func EnsureCheckedIn(c *srv.Service) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(model.WhimerSessId)
			if err != nil || len(cookie.Value) == 0 {
				httpx.Error(w, global.ErrNotCheckedIn)
				return
			}

			sessId, err := url.PathUnescape(cookie.Value)
			if err != nil {
				httpx.Error(w, global.ErrArgs.Msg(err.Error()))
				return
			}

			// 获取sessid的信息
			user, err := c.AccessSrv.IsCheckedIn(r.Context(), sessId)
			if err != nil {
				httpx.Error(w, err)
				return
			}

			// 注入后续使用的参数
			ctx := model.WithUserInfo(r.Context(), user)
			ctx = model.WithSessId(ctx, sessId)
			nr := r.WithContext(ctx)

			next(w, nr)
		}
	}
}
