package global

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/errorx"
)

const (
	PassportErrCode = 30000
)

var (
	ErrArgs       = errorx.ErrInvalidArgs.ErrCode(PassportErrCode).Msg("参数错误")
	ErrInternal   = errorx.ErrInternal.ErrCode(PassportErrCode).Msg("服务错误, 请稍后重试")
	ErrPermDenied = errorx.ErrPermission.ErrCode(PassportErrCode).Msg("操作权限不足")
	ErrRateLimit  = errorx.NewError(http.StatusTooManyRequests, PassportErrCode, "你的操作太频繁了")

	// sign-in related
	ErrRegisterTel        = ErrInternal.Msg("注册失败")
	ErrRequestSms         = ErrInternal.Msg("获取验证码失败, 请稍后重试")
	ErrRequestSmsFrequent = ErrRateLimit.Msg("短信请求过快, 请稍后重试")
	ErrUserNotRegister    = ErrPermDenied.Msg("你还没有注册")
	ErrSmsCodeNotMatch    = ErrPermDenied.Msg("验证码错误")
	ErrPassNotMatch       = ErrPermDenied.Msg("密码不对")
	ErrInvalidTel         = ErrArgs.Msg("手机号格式不正确")

	// sign-up related
	ErrTelTaken   = ErrPermDenied.Msg("该手机号已经注册")
	ErrEmailTaken = ErrPermDenied.Msg("该邮箱已经使用")
)
