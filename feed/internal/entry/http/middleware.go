package http

import (
	"github.com/ryanreadbooks/whimer/feed/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/passport/pkg/middleware/auth"
	
	"github.com/zeromicro/go-zero/rest"
)

// 必须登录
func MustLogin() rest.Middleware {
	return auth.UserWeb(dep.Auther())
}

// 可以不用登录 也可以登录
func CanLogin() rest.Middleware {
	return auth.UserWebOptional(dep.Auther())
}
