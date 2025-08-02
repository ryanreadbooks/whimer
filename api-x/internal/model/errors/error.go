package errors

import "github.com/ryanreadbooks/whimer/misc/xerror"

var (
	ErrNoteNotFound =  xerror.ErrArgs.Msg("笔记不存在")
)