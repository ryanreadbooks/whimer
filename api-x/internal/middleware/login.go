package middleware

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/backend/passport"
	"github.com/ryanreadbooks/whimer/passport/sdk/middleware/auth"
	"github.com/zeromicro/go-zero/rest"
)

// 必须登录
func MustLogin() rest.Middleware {
	return auth.UserWeb(passport.Auther())
}

// 可以不用登录 也可以登录
func CanLogin() rest.Middleware {
	return auth.UserWebOptional(passport.Auther())
}
