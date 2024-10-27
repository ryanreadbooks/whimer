package global

import "github.com/ryanreadbooks/whimer/misc/xerror"

const (
	CounterErrCode = xerror.BizCounter
)

// 业务错误定义
var (
	ErrBizArgs     = xerror.ErrInvalidArgs.ErrCode(CounterErrCode)
	ErrBizInternal = xerror.ErrInternal.ErrCode(CounterErrCode)
	ErrBizDenied   = xerror.ErrPermission.ErrCode(CounterErrCode)
	ErrNotFound    = xerror.ErrNotFound.ErrCode(CounterErrCode)

	ErrArgs         = ErrBizArgs.Msg("参数错误")
	ErrInternal     = ErrBizInternal.Msg("服务错误, 请稍后重试")
	ErrPermDenied   = ErrBizDenied.Msg("操作权限不足")
	ErrNilReq       = ErrArgs.Msg("请求参数为空")
	ErrNoRecord     = ErrNotFound.Msg("找不到记录")
	ErrAlreadyDo    = ErrArgs.Msg("不能重复操作")
	ErrCountSummary = ErrInternal.Msg("获取计数失败，请稍后重试")
)
