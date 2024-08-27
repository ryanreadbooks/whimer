package errorx

import (
	"encoding/json"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Error struct {
	StatusCode int    `json:"-"`    // http响应状态码
	Code       int    `json:"code"` // 业务响应码
	Message    string `json:"msg"`
}

func (e *Error) Error() string {
	s, _ := json.Marshal(e)
	return utils.Bytes2String(s)
}

func (e *Error) Msg(msg string) *Error {
	if e == nil {
		return nil
	}

	return &Error{
		Code:       e.Code,
		StatusCode: e.StatusCode,
		Message:    msg,
	}
}

func (e *Error) ExtMsg(extmsg string) *Error {
	// 保留原来msg的基础下 在msg中新增extmsg
	if e == nil {
		return nil
	}

	msg := e.Message + ": " + extmsg

	return &Error{
		Code:       e.Code,
		StatusCode: e.StatusCode,
		Message:    msg,
	}
}

func (e *Error) ErrCode(ecode int) *Error {
	if e == nil {
		return nil
	}

	return &Error{
		Code:       ecode,
		StatusCode: e.StatusCode,
		Message:    e.Message,
	}
}

func NewError(st, code int, msg string) *Error {
	return &Error{
		StatusCode: st,
		Code:       code,
		Message:    msg,
	}
}

func IsInternal(err error) bool {
	if e, ok := err.(*Error); ok {
		return e.Code >= http.StatusInternalServerError
	}

	return true
}

// 判断给定的err是否应该打印日志
func ShouldLog(err error) bool {
	if err == nil {
		return false
	}

	// 判断是否为Error对象
	commErr, ok := err.(*Error)
	if ok {
		return commErr.StatusCode >= http.StatusInternalServerError
	}

	// 判断是否为grpc.Status
	grpcerr, ok := status.FromError(err)
	if ok {
		switch grpcerr.Code() {
		case codes.OK,
			codes.NotFound,
			codes.InvalidArgument,
			codes.AlreadyExists,
			codes.PermissionDenied,
			codes.Unauthenticated:
			return false
		default:
			return true
		}
	}

	return false
}
