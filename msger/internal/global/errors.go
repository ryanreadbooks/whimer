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

	ErrNilReq = ErrArgs.Msg("请求参数为空")

	ErrChatMsgNil           = ErrArgs.Msg("发送消息为空")
	ErrP2PChatNotExist      = ErrArgs.Msg("会话不存在")
	ErrP2PChatUserEmpty     = ErrArgs.Msg("会话用户不存在")
	ErrP2PChatSenderEmpty   = ErrArgs.Msg("发送者不存在")
	ErrP2PChatReceiverEmpty = ErrArgs.Msg("接收者不存在")
	ErrUserNotInChat        = ErrArgs.Msg("用户不在会话中")
	ErrMsgAlreadyRevoked    = ErrArgs.Msg("消息已被撤回")
	ErrMsgNotExist          = ErrArgs.Msg("消息不存在")
	ErrCantRevokeMsg        = ErrPermDenied.Msg("无权撤回该消息")
)
