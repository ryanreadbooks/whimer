package errors

import "github.com/ryanreadbooks/whimer/misc/xerror"

var (
	ErrTooManyNotes       = xerror.ErrInvalidArgs.Msg("不能拿这么多")
	ErrNoteNotFound       = xerror.ErrInvalidArgs.Msg("笔记不存在")
	ErrUserNotFound       = xerror.ErrInvalidArgs.Msg("用户不存在")
	ErrLikesHistoryHidden = xerror.ErrInvalidArgs.Msg("点赞记录隐藏")
)
