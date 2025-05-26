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

	ErrP2PChatNotExist = ErrArgs.Msg("会话不存在")
)
