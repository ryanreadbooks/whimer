package global

import "github.com/ryanreadbooks/whimer/misc/errorx"

const (
	CounterErrCode = errorx.BizCounter
)

// 业务错误定义
var (
	ErrBizArgs     = errorx.ErrInvalidArgs.ErrCode(CounterErrCode)
	ErrBizInternal = errorx.ErrInternal.ErrCode(CounterErrCode)
	ErrBizDenied   = errorx.ErrPermission.ErrCode(CounterErrCode)
	ErrNotFound    = errorx.ErrNotFound.ErrCode(CounterErrCode)

	ErrArgs         = ErrBizArgs.Msg("参数错误")
	ErrInternal     = ErrBizInternal.Msg("服务错误, 请稍后重试")
	ErrPermDenied   = ErrBizDenied.Msg("操作权限不足")
	ErrNilReq       = ErrArgs.Msg("请求参数为空")
	ErrNoRecord     = ErrNotFound.Msg("找不到记录")
	ErrAlreadyDo    = ErrArgs.Msg("不能重复操作")
	ErrCountSummary = ErrInternal.Msg("获取计数失败，请稍后重试")
)
