package global

import "github.com/ryanreadbooks/whimer/misc/xerror"

const (
	MsgerErrCode = xerror.BizMsger
)

// 业务错误定义
var (
	ErrBizMsgerArgs     = xerror.ErrInvalidArgs.ErrCode(MsgerErrCode)
	ErrBizMsgerInternal = xerror.ErrInternal.ErrCode(MsgerErrCode)
	ErrBizMsgerDenied   = xerror.ErrPermission.ErrCode(MsgerErrCode)
	ErrNotFound         = xerror.ErrNotFound.ErrCode(MsgerErrCode)

	ErrArgs       = ErrBizMsgerArgs.Msg("参数错误")
	ErrInternal   = ErrBizMsgerInternal.Msg("服务错误, 请稍后重试")
	ErrPermDenied = ErrBizMsgerDenied.Msg("操作权限不足")

	ErrNilReq      = ErrArgs.Msg("请求参数为空")
	ErrLockNotHeld = ErrArgs.Msg("请稍后重试")

	ErrChatMsgNil            = ErrArgs.Msg("发送消息为空")
	ErrChatNotExist          = ErrArgs.Msg("会话不存在")
	ErrChatUserEmpty         = ErrArgs.Msg("会话用户不存在")
	ErrChatSenderEmpty       = ErrArgs.Msg("发送者不存在")
	ErrChatReceiverEmpty     = ErrArgs.Msg("接收者不存在")
	ErrUserNotInChat         = ErrArgs.Msg("用户不在会话中")
	ErrMsgAlreadyRevoked     = ErrArgs.Msg("消息已被撤回")
	ErrMsgRevokedTimeReached = ErrArgs.Msg("超过撤回时间")
	ErrMsgNotExist           = ErrArgs.Msg("消息不存在")
	ErrCantRevokeMsg         = ErrPermDenied.Msg("无权撤回该消息")
	ErrUserNotFound          = ErrArgs.Msg("用户不存在")
	ErrUnsupportedMsgType    = ErrArgs.Msg("不支持的消息类型")
	ErrUnsupportedChatType   = ErrArgs.Msg("不支持的会话类型")
	ErrEmptyMsg              = ErrArgs.Msg("消息内容为空")
	ErrSysChatNotExist       = ErrArgs.Msg("系统会话不存在")
	ErrSysChatNotYours       = ErrPermDenied.Msg("系统消息归属错误")
	ErrGenChatId             = ErrInternal.Msg("无法生成会话id")
	ErrChatNotNormal         = ErrArgs.Msg("会话状态异常")
)
