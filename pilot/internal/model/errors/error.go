package errors

import "github.com/ryanreadbooks/whimer/misc/xerror"

// 400
var (
	ErrNoteNotFound          = xerror.ErrArgs.Msg("笔记不存在")
	ErrTagNotFound           = xerror.ErrArgs.Msg("标签不存在")
	ErrUserNotFound          = xerror.ErrArgs.Msg("用户不存在")
	ErrReplyUserDoesNotMatch = xerror.ErrArgs.Msg("回复用户错误")
	ErrUnsupportedChatType   = xerror.ErrArgs.Msg("不支持的会话类型")
	ErrUnsupportedMsgType    = xerror.ErrArgs.Msg("不支持的消息类型")
	ErrInvalidMsgContent     = xerror.ErrArgs.Msg("无效消息内容")
	ErrChatNotExists         = xerror.ErrArgs.Msg("会话不存在")
	ErrChatMsgNotExists      = xerror.ErrArgs.Msg("消息不存在")
	ErrUnsupportedResource   = xerror.ErrArgs.Msg("不支持的资源类型")
)

// 5xx
var (
	ErrServerSignFailure = xerror.ErrInternal.Msg("服务器签名失败")
)

// 403
var (
	ErrLikesHistoryHidden = xerror.ErrPermission.Msg("用户隐藏了点赞记录")
)
