package errors

import "github.com/ryanreadbooks/whimer/misc/xerror"

var (
	ErrUserNotFound        = xerror.ErrArgs.Msg("用户不存在")
	ErrUnsupportedChatType = xerror.ErrArgs.Msg("不支持的会话类型")
	ErrUnsupportedMsgType  = xerror.ErrArgs.Msg("不支持的消息类型")
	ErrInvalidMsgContent   = xerror.ErrArgs.Msg("无效消息内容")
	ErrChatNotExists       = xerror.ErrArgs.Msg("会话不存在")
	ErrChatMsgNotExists    = xerror.ErrArgs.Msg("消息不存在")
	ErrInvalidOrder        = xerror.ErrArgs.Msg("无效的排序方式")
)
