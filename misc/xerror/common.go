package xerror

import (
	"encoding/json"
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Error struct {
	StatusCode int    `json:"stcode,omitempty"` // http响应状态码
	Code       int    `json:"code"`             // 业务响应码
	Message    string `json:"msg"`
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}

	ne := *e
	ne.StatusCode = 0 // 不对外
	return ne.Json()
}

func (e *Error) Json() string {
	if e == nil {
		return ""
	}
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

func (e *Error) Is(err error) bool {
	if oth, ok := err.(*Error); ok {
		return e.Equal(oth)
	}
	return false
}

func (e *Error) Equal(other *Error) bool {
	if other == nil {
		return e == nil
	}
	if e == nil {
		return other == nil
	}

	// 不要求Msg相等
	return e.Code == other.Code && e.StatusCode == other.StatusCode
}

func (e *Error) ShouldLogError() bool {
	return e.Code >= http.StatusInternalServerError
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
func ShouldLogError(err error) bool {
	if err == nil {
		return false
	}

	// 有限检查是否为errProxy
	stackErr, ok := err.(*errProxy)
	var (
		commErr  *Error
		causeErr error = err
	)

	if ok {
		// 判断底层是否为Error对象
		cause := Cause(stackErr)
		commErr, ok = cause.(*Error)
		if !ok {
			causeErr = cause
		}
	} else {
		// 判断是否为Error对象
		commErr, _ = err.(*Error)
	}

	if commErr != nil {
		return commErr.StatusCode >= http.StatusInternalServerError
	}

	// 判断是否为grpc.Status
	grpcerr, ok := status.FromError(causeErr)
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

func FromJson(c string) *Error {
	var e Error
	err := json.Unmarshal(utils.StringToBytes(c), &e)
	if err != nil {
		return ErrOther
	}

	return &e
}
