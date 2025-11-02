package global

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xerror"
)

const (
	PassportErrCode = xerror.BizPassport

	ErrInvalidArgsCode = PassportErrCode + iota*1000
	ErrInternalCode
	ErrPermissionCode
	ErrNotLoginCode
	ErrRateLimitCode
	ErrMethodNotAllowedCode
)

const (
	_ = iota

	ErrPassportNilReqCode = ErrInvalidArgsCode + iota
	ErrPassportInvalidTelCode
	ErrPassportInvalidSmsCodeCode
	ErrPassportInvalidPlatformCode
	ErrPassportInvalidUidCode
	ErrPassportNicknameTooShortCode
	ErrPassportNickNameTooLongCode
	ErrPassportInvalidGenderCode
	ErrPassportStyleSignTooLongCode
	ErrPassportReadFileCode
	ErrPassportAvatarNotFoundCode
	ErrPassportAvatarTooLargeCode
	ErrPassportAvatarFormatUnsupportedCode
)

const (
	_ = iota

	ErrPassportRegisterTelCode = ErrInternalCode + iota
	ErrPassportRequestSmsCode
	ErrPassportCheckInCode
	ErrPassportAccessBizCode
	ErrPassportSessRenewalCode
	ErrPassportUploadAvatarCode
	ErrPassportGetUserFailCode
)

const (
	_ = iota

	ErrPassportUserNotRegisterCode = ErrPermissionCode + iota
	ErrPassportSmsCodeNotMatchCode
	ErrPassportPassNotMatchCode
	ErrPassportSessInvalidatedCode
	ErrPassportCheckOutCode
	ErrPassportSessPlatformNotMatchedCode
	ErrPassportUserNotFoundCode
	ErrPassportTelTakenCode
	ErrPassportEmailTakenCode
	ErrPassportNicknameTakenCode
)

const (
	_ = iota

	ErrPassportRequestSmsFrequentCode = ErrRateLimitCode + iota
	ErrPassportCheckInTooFrequentCode
)

var (
	ErrArgs             = xerror.ErrInvalidArgs.ErrCode(ErrInvalidArgsCode).Msg("参数错误")
	ErrInternal         = xerror.ErrInternal.ErrCode(ErrInternalCode).Msg("服务错误, 请稍后重试")
	ErrUnAuth           = xerror.ErrNotLogin.ErrCode(ErrNotLoginCode)
	ErrPermDenied       = xerror.ErrPermission.ErrCode(ErrPermissionCode).Msg("操作权限不足")
	ErrRateLimit        = xerror.NewError(http.StatusTooManyRequests, ErrRateLimitCode, "你的操作太频繁了")
	ErrApiUnimplemented = xerror.NewError(http.StatusMethodNotAllowed, ErrMethodNotAllowedCode, "接口未实现")
	ErrNilReq           = xerror.ErrInvalidArgs.ErrCode(ErrPassportNilReqCode).Msg("请求为空")

	// sign-in related
	ErrNotCheckedIn            = ErrUnAuth
	ErrRegisterTel             = xerror.ErrInternal.ErrCode(ErrPassportRegisterTelCode).Msg("注册失败")
	ErrRequestSms              = xerror.ErrInternal.ErrCode(ErrPassportRequestSmsCode).Msg("获取验证码失败, 请稍后重试")
	ErrRequestSmsFrequent      = xerror.NewError(http.StatusTooManyRequests, ErrPassportRequestSmsFrequentCode, "短信请求过快, 请稍后重试")
	ErrUserNotRegister         = xerror.ErrPermission.ErrCode(ErrPassportUserNotRegisterCode).Msg("你还没有注册")
	ErrSmsCodeNotMatch         = xerror.ErrPermission.ErrCode(ErrPassportSmsCodeNotMatchCode).Msg("验证码错误")
	ErrPassNotMatch            = xerror.ErrPermission.ErrCode(ErrPassportPassNotMatchCode).Msg("密码不对")
	ErrInvalidTel              = xerror.ErrInvalidArgs.ErrCode(ErrPassportInvalidTelCode).Msg("手机号格式不正确")
	ErrInvalidSmsCode          = xerror.ErrInvalidArgs.ErrCode(ErrPassportInvalidSmsCodeCode).Msg("短信验证码格式不正确")
	ErrInvalidPlatform         = xerror.ErrInvalidArgs.ErrCode(ErrPassportInvalidPlatformCode).Msg("暂不支持该平台")
	ErrCheckIn                 = xerror.ErrInternal.ErrCode(ErrPassportCheckInCode).Msg("登录失败, 请稍后重试")
	ErrAccessBiz               = xerror.ErrInternal.ErrCode(ErrPassportAccessBizCode).Msg("服务异常，请稍后重试")
	ErrSessRenewal             = xerror.ErrInternal.ErrCode(ErrPassportSessRenewalCode).Msg("续期失败, 请稍后重试")
	ErrCheckInTooFrequent      = xerror.NewError(http.StatusTooManyRequests, ErrPassportCheckInTooFrequentCode, "登录太快, 请稍后重试")
	ErrSessInvalidated         = xerror.ErrPermission.ErrCode(ErrPassportSessInvalidatedCode).Msg("登录过期, 请重新登录")
	ErrCheckOut                = xerror.ErrPermission.ErrCode(ErrPassportCheckOutCode).Msg("退出登录失败，请稍后重试")
	ErrSessPlatformNotMatched  = xerror.ErrPermission.ErrCode(ErrPassportSessPlatformNotMatchedCode).Msg("未在此设备上登录")
	ErrUserNotFound            = xerror.ErrPermission.ErrCode(ErrPassportUserNotFoundCode).Msg("没有找到你的信息")
	ErrInvalidUid              = xerror.ErrInvalidArgs.ErrCode(ErrPassportInvalidUidCode).Msg("uid格式不对")
	ErrNicknameTooShort        = xerror.ErrInvalidArgs.ErrCode(ErrPassportNicknameTooShortCode).Msg("昵称不能为空")
	ErrNickNameTooLong         = xerror.ErrInvalidArgs.ErrCode(ErrPassportNickNameTooLongCode).Msg("你的昵称太长啦")
	ErrInvalidGender           = xerror.ErrInvalidArgs.ErrCode(ErrPassportInvalidGenderCode).Msg("没有这种性别")
	ErrStyleSignTooLong        = xerror.ErrInvalidArgs.ErrCode(ErrPassportStyleSignTooLongCode).Msg("你的个性签名太长啦")
	ErrReadFile                = xerror.ErrInvalidArgs.ErrCode(ErrPassportReadFileCode).Msg("读取失败")
	ErrUploadAvatar            = xerror.ErrInternal.ErrCode(ErrPassportUploadAvatarCode).Msg("上传头像失败")
	ErrAvatarNotFound          = xerror.ErrInvalidArgs.ErrCode(ErrPassportAvatarNotFoundCode).Msg("没有图像文件")
	ErrAvatarTooLarge          = xerror.ErrInvalidArgs.ErrCode(ErrPassportAvatarTooLargeCode).Msg("上传的头像太大")
	ErrAvatarFormatUnsupported = xerror.ErrInvalidArgs.ErrCode(ErrPassportAvatarFormatUnsupportedCode).Msg("不支持的头像格式")

	// sign-up related
	ErrTelTaken      = xerror.ErrPermission.ErrCode(ErrPassportTelTakenCode).Msg("该手机号已经注册")
	ErrEmailTaken    = xerror.ErrPermission.ErrCode(ErrPassportEmailTakenCode).Msg("该邮箱已经使用")
	ErrNicknameTaken = xerror.ErrPermission.ErrCode(ErrPassportNicknameTakenCode).Msg("该昵称已被占用")

	// user profile
	ErrGetUserFail = xerror.ErrInternal.ErrCode(ErrPassportGetUserFailCode).Msg("获取用户信息失败")
)
