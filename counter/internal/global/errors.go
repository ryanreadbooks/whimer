package global

import "github.com/ryanreadbooks/whimer/misc/xerror"

const (
	ErrCounterCode = xerror.BizCounter

	// Counter错误码常量定义 - 使用iota

	ErrInvalidArgsCode = ErrCounterCode + iota*1000
	ErrInternalCode
	ErrPermissionCode
	ErrNotFoundCode
)

const (
	_ = iota

	ErrCounterNilReqCode = ErrInvalidArgsCode + iota
	ErrCounterAlreadyDoCode
)

const (
	_ = iota

	ErrCounterNoRecordCode = ErrNotFoundCode + iota
)

const (
	_ = iota

	ErrCounterCountSummaryCode = ErrInternalCode + iota
)

// 业务错误定义
var (
	ErrBizArgs     = xerror.ErrInvalidArgs.ErrCode(ErrInvalidArgsCode)
	ErrBizInternal = xerror.ErrInternal.ErrCode(ErrInternalCode)
	ErrBizDenied   = xerror.ErrPermission.ErrCode(ErrPermissionCode)
	ErrNotFound    = xerror.ErrNotFound.ErrCode(ErrNotFoundCode)

	ErrArgs         = ErrBizArgs.Msg("参数错误")
	ErrInternal     = ErrBizInternal.Msg("服务错误, 请稍后重试")
	ErrPermDenied   = ErrBizDenied.Msg("操作权限不足")
	ErrNilReq       = ErrBizArgs.ErrCode(ErrCounterNilReqCode).Msg("请求参数为空")
	ErrNoRecord     = ErrNotFound.ErrCode(ErrCounterNoRecordCode).Msg("找不到记录")
	ErrAlreadyDo    = ErrBizArgs.ErrCode(ErrCounterAlreadyDoCode).Msg("不能重复操作")
	ErrCountSummary = ErrBizInternal.ErrCode(ErrCounterCountSummaryCode).Msg("获取计数失败，请稍后重试")
)
