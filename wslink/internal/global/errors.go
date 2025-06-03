package global

import "github.com/ryanreadbooks/whimer/misc/xerror"

const (
	WsErrCode = 70000
)

// 业务错误定义
var (
	ErrBizArgs     = xerror.ErrInvalidArgs.ErrCode(WsErrCode)
	ErrBizInternal = xerror.ErrInternal.ErrCode(WsErrCode)
	ErrBizDenied   = xerror.ErrPermission.ErrCode(WsErrCode)
	ErrNotFound    = xerror.ErrNotFound.ErrCode(WsErrCode)

	ErrAuthFailed = xerror.ErrPermission.Msg("认证失败")
	ErrServerBusy = xerror.ErrServiceUnavailable.Msg("系统繁忙，稍后重试")
)
