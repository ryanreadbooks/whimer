package errorx

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
	CodeOther              = 19999
)

// 常用错误
var (
	ErrSuccess            = NewError(http.StatusOK, CodeOk, "成功")
	ErrInvalidArgs        = NewError(BadRequest, CodeInvalidParam, "参数出错了(；一_一)")
	ErrArgs               = ErrInvalidArgs
	ErrNilArg             = ErrArgs.Msg("参数为空")
	ErrNotLogin           = NewError(Unauthorized, CodeUnauthorized, "请先登录一下吧~(≧▽≦)")
	ErrPermission         = NewError(Forbidden, CodeForbidden, "需要特殊通行证")
	ErrNotFound           = NewError(NotFound, CodeNotFound, "找不到你要的资源")
	ErrInternal           = NewError(InternalServerError, CodeInternal, "服务器被怪兽踢烂了(ノ｀Д´)ノ")
	ErrCsrf               = NewError(Forbidden, CodeCsrfFailed, "CSRF校验失败")
	ErrServiceUnavailable = NewError(ServiceUnavailable, CodeServiceUnavailable, "服务暂不可用")
)
