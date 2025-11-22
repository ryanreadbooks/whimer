package global

import "github.com/ryanreadbooks/whimer/misc/xerror"

const (
	WsErrCode = 70000

	// Wslink error groups

	ErrInvalidArgsCode = WsErrCode + iota*1000
	ErrInternalCode
	ErrPermissionCode
	ErrNotFoundCode
	ErrServiceUnavailableCode
)

const (
	_ = iota

	ErrWsUserEmptyCode = ErrInvalidArgsCode + iota
	ErrWsUnsupportedDeviceCode
	ErrWsDataEmptyCode
	ErrReqIdMissingCode
)

const (
	_ = iota

	ErrWsAuthFailedCode = ErrPermissionCode + iota
)

const (
	_ = iota

	ErrWsServerBusyCode = ErrServiceUnavailableCode + iota
)

// 业务错误定义
var (
	ErrBizArgs     = xerror.ErrInvalidArgs.ErrCode(ErrInvalidArgsCode)
	ErrBizInternal = xerror.ErrInternal.ErrCode(ErrInternalCode)
	ErrBizDenied   = xerror.ErrPermission.ErrCode(ErrPermissionCode)
	ErrNotFound    = xerror.ErrNotFound.ErrCode(ErrNotFoundCode)

	ErrUserEmpty         = xerror.ErrInvalidArgs.ErrCode(ErrWsUserEmptyCode).Msg("用户id非法")
	ErrUnsupportedDevice = xerror.ErrInvalidArgs.ErrCode(ErrWsUnsupportedDeviceCode).Msg("不支持的设备")
	ErrDataEmpty         = xerror.ErrInvalidArgs.ErrCode(ErrWsDataEmptyCode).Msg("内容为空")
	ErrAuthFailed        = xerror.ErrPermission.ErrCode(ErrWsAuthFailedCode).Msg("认证失败")
	ErrServerBusy        = xerror.ErrServiceUnavailable.ErrCode(ErrWsServerBusyCode).Msg("系统繁忙，稍后重试")
	ErrReqIdMissing      = xerror.ErrInvalidArgs.ErrCode(ErrReqIdMissingCode).Msg("reqId missing")
)
