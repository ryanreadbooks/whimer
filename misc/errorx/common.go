package errorx

import (
	"encoding/json"

	"github.com/ryanreadbooks/whimer/misc/utils"
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
