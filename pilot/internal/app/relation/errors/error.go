package errors

import "github.com/ryanreadbooks/whimer/misc/xerror"

var (
	ErrInvalidUserId = xerror.ErrInvalidArgs.Msg("非法用户id")
	ErrInvalidAction = xerror.ErrInvalidArgs.Msg("不支持的操作")
	ErrUserNotFound  = xerror.ErrInvalidArgs.Msg("用户不存在")
)
