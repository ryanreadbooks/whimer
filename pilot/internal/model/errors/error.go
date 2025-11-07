package errors

import "github.com/ryanreadbooks/whimer/misc/xerror"

var (
	ErrNoteNotFound          = xerror.ErrArgs.Msg("笔记不存在")
	ErrTagNotFound           = xerror.ErrArgs.Msg("标签不存在")
	ErrUserNotFound          = xerror.ErrArgs.Msg("用户不存在")
	ErrReplyUserDoesNotMatch = xerror.ErrArgs.Msg("回复用户错误")
	ErrUnsupportedChatType   = xerror.ErrArgs.Msg("不支持的会话类型")
)
