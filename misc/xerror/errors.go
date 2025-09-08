package xerror

import "net/http"

// 业务错误码定义
const (
	CodeOk                 = 0
	CodeUnauthorized       = 10001
	CodeInvalidParam       = 10002
	CodeForbidden          = 10003
	CodeNotFound           = 10004
	CodeInternal           = 10005
	CodeCsrfFailed         = 10006
	CodeServiceUnavailable = 10007
	CodeDuplicate          = 10008
	CodeTooLong            = 10009
	CodeInternalPanic      = 19000
	CodeOther              = 19999
)

// 常用错误
var (
	Success               = NewError(http.StatusOK, CodeOk, "成功")
	ErrInvalidArgs        = NewError(BadRequest, CodeInvalidParam, "参数出错了(；一_一)")
	ErrArgs               = ErrInvalidArgs
	ErrNilArg             = ErrArgs.Msg("参数为空")
	ErrNotLogin           = NewError(Unauthorized, CodeUnauthorized, "请先登录一下吧~(≧▽≦)")
	ErrPermission         = NewError(Forbidden, CodeForbidden, "需要特殊通行证")
	ErrNotFound           = NewError(NotFound, CodeNotFound, "找不到你要的资源")
	ErrInternal           = NewError(InternalServerError, CodeInternal, "服务器被怪兽踢烂了(ノ｀Д´)ノ")
	ErrDepNotReady        = NewError(InternalServerError, CodeInternal, "服务未就绪")
	ErrInternalPanic      = NewError(InternalServerError, CodeInternalPanic, "服务器炸掉了")
	ErrDuplicate          = NewError(InternalServerError, CodeDuplicate, "资源重复")
	ErrDataTooLong        = NewError(InternalServerError, CodeTooLong, "数据超长")
	ErrCsrf               = NewError(Forbidden, CodeCsrfFailed, "CSRF校验失败")
	ErrServiceUnavailable = NewError(ServiceUnavailable, CodeServiceUnavailable, "服务暂不可用")
	ErrOther              = NewError(InternalServerError, CodeOther, "服务错误")
)

// internal user
var (
	ErrApiWentOffline = NewError(NotFound, CodeOther, "接口不存在(´･ω･`)")
	ErrPanic          = NewError(InternalServerError, -99999, "FATAL PANIC")
)
