package errorx

import "net/http"

// 常用错误
var (
	ErrSuccess = NewError(http.StatusOK, 0, "success")

	ErrInvalidArgs = NewError(http.StatusBadRequest, CommonCode, "请求参数错误")
	ErrNotLogin    = NewError(http.StatusUnauthorized, CommonCode, "请先登录")
	ErrPermission  = NewError(http.StatusForbidden, CommonCode, "权限不足")
	ErrNotFound    = NewError(http.StatusNotFound, CommonCode, "资源不存在")
	ErrInternal    = NewError(http.StatusInternalServerError, CommonCode, "服务出错, 稍后再试")
)

// 业务错误码定义
const (
	CommonCode = 10000
)
