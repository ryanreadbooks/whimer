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
	ErrUnAuth     = errorx.ErrNotLogin.ErrCode(PassportErrCode)
	ErrPermDenied = errorx.ErrPermission.ErrCode(PassportErrCode).Msg("操作权限不足")
	ErrRateLimit  = errorx.NewError(http.StatusTooManyRequests, PassportErrCode, "你的操作太频繁了")

	// sign-in related
	ErrNotSignedIn             = ErrUnAuth
	ErrRegisterTel             = ErrInternal.Msg("注册失败")
	ErrRequestSms              = ErrInternal.Msg("获取验证码失败, 请稍后重试")
	ErrRequestSmsFrequent      = ErrRateLimit.Msg("短信请求过快, 请稍后重试")
	ErrUserNotRegister         = ErrPermDenied.Msg("你还没有注册")
	ErrSmsCodeNotMatch         = ErrPermDenied.Msg("验证码错误")
	ErrPassNotMatch            = ErrPermDenied.Msg("密码不对")
	ErrInvalidTel              = ErrArgs.Msg("手机号格式不正确")
	ErrInvalidSmsCode          = ErrArgs.Msg("短信验证码格式不正确")
	ErrInvalidPlatform         = ErrArgs.Msg("暂不支持该平台")
	ErrSignIn                  = ErrInternal.Msg("登录失败, 请稍后重试")
	ErrSessRenewal             = ErrInternal.Msg("续期失败, 请稍后重试")
	ErrSignInTooFrequent       = ErrRateLimit.Msg("登录太快, 请稍后重试")
	ErrSessInvalidated         = ErrPermDenied.Msg("登录过期, 请重新登录")
	ErrSessPlatformNotMatched  = ErrPermDenied.Msg("未在此设备上登录")
	ErrMeNotFound              = ErrPermDenied.Msg("没有找到你的信息")
	ErrInvalidUid              = ErrArgs.Msg("uid格式不对")
	ErrNicknameTooShort        = ErrArgs.Msg("昵称不能为空")
	ErrNickNameTooLong         = ErrArgs.Msg("你的昵称太长啦")
	ErrInvalidGender           = ErrArgs.Msg("没有这种性别")
	ErrStyleSignTooLong        = ErrArgs.Msg("你的个性签名太长啦")
	ErrReadFile                = ErrArgs.Msg("读取失败")
	ErrUploadAvatar            = ErrInternal.Msg("上传头像失败")
	ErrAvatarNotFound          = ErrArgs.Msg("没有图像文件")
	ErrAvatarTooLarge          = ErrArgs.Msg("上传的头像太大")
	ErrAvatarFormatUnsupported = ErrArgs.Msg("不支持的头像格式")

	// sign-up related
	ErrTelTaken      = ErrPermDenied.Msg("该手机号已经注册")
	ErrEmailTaken    = ErrPermDenied.Msg("该邮箱已经使用")
	ErrNicknameTaken = ErrPermDenied.Msg("该昵称已被占用")
)
