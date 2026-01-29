package errors

import "github.com/ryanreadbooks/whimer/misc/xerror"

var (
	ErrUnsupportedAction = xerror.ErrArgs.Msg("不支持的操作")
)