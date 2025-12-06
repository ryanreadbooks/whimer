package global

import "github.com/ryanreadbooks/whimer/misc/xerror"

const (
	ConductorErrCode = xerror.BizConductor

	// Conductor error groups

	ErrInvalidArgsCode = ConductorErrCode + iota*1000
	ErrInternalCode
	ErrPermissionCode
	ErrNotFoundCode
)

const (
	_ = iota

	ErrNamespaceAlreadyExistsCode = ErrInvalidArgsCode + iota
	ErrNamespaceNotFoundCode
)

var (
	ErrBizArgs     = xerror.ErrInvalidArgs.ErrCode(ErrInvalidArgsCode)
	ErrBizInternal = xerror.ErrInternal.ErrCode(ErrInternalCode)
	ErrBizDenied   = xerror.ErrPermission.ErrCode(ErrPermissionCode)
	ErrNotFound    = xerror.ErrNotFound.ErrCode(ErrNotFoundCode)

	ErrArgs                   = ErrBizArgs.Msg("参数错误")
	ErrInternal               = ErrBizInternal.Msg("服务错误, 请稍后重试")
	ErrPermDenied             = ErrBizDenied.Msg("操作权限不足")
	ErrNamespaceAlreadyExists = ErrBizArgs.ErrCode(ErrNamespaceAlreadyExistsCode).Msg("命名空间已存在")
	ErrNamespaceNotFound      = ErrBizArgs.ErrCode(ErrNamespaceNotFoundCode).Msg("命名空间不存在")
)
