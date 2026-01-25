package errors

import "github.com/ryanreadbooks/whimer/misc/xerror"

var (
	ErrInvalidUserId = xerror.ErrArgs.Msg("invalid user id")
	ErrUserNotFound  = xerror.ErrNotFound.Msg("user not found")
)
