package errorx

import "net/http"

// 业务错误码定义
const (
	CommonCodeOk           = 0
	CommonCodeUnauthorized = 10001
	CommonCodeInvalidParam = 10002
	CommonCodeForbidden    = 10003
	CommonCodeNotFound     = 10004
	CommonCodeInternal     = 10005
)

// 常用错误
var (
	ErrSuccess = NewError(http.StatusOK, CommonCodeOk, "success")

	ErrInvalidArgs = NewError(http.StatusBadRequest, CommonCodeInvalidParam, "参数出错了(；一_一)")
	ErrNotLogin    = NewError(http.StatusUnauthorized, CommonCodeUnauthorized, "请先登录一下吧~(≧▽≦)")
	ErrPermission  = NewError(http.StatusForbidden, CommonCodeForbidden, "需要特殊通行证")
	ErrNotFound    = NewError(http.StatusNotFound, CommonCodeNotFound, "找不到你要的资源")
	ErrInternal    = NewError(http.StatusInternalServerError, CommonCodeInternal, "服务器被怪兽踢烂了(ノ｀Д´)ノ")
)
