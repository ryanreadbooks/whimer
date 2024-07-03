package global

import "github.com/ryanreadbooks/whimer/misc/errorx"

const (
	PassportErrCode = 30000
)

var (
	ErrArgs       = errorx.ErrInvalidArgs.ErrCode(PassportErrCode).Msg("参数错误")
	ErrInternal   = errorx.ErrInternal.ErrCode(PassportErrCode).Msg("笔记服务错误, 请稍后重试")
	ErrPermDenied = errorx.ErrPermission.ErrCode(PassportErrCode).Msg("操作权限不足")

	ErrRegisterTel = ErrInternal.Msg("注册失败")
	ErrTelTaken    = ErrPermDenied.Msg("该手机号已经注册")
	ErrEmailTaken  = ErrPermDenied.Msg("该邮箱已经使用")
)
