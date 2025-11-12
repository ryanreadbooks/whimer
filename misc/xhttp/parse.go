package xhttp

import (
	"net/http"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/zeromicro/go-zero/rest/httpx"
)

type Validator interface {
	Validate() error
}

type ParserFunc func(*http.Request, any) error

// 解析请求并执行校验动作
// 返回解析后的请求对象
func ParseValidate[T any](parser ParserFunc, r *http.Request) (out *T, err error) {
	t := new(T)
	if err := parser(r, t); err != nil {
		return nil, xerror.ErrArgs.Msg(err.Error())
	}

	if validator, ok := any(t).(Validator); ok && validator != nil {
		if err := validator.Validate(); err != nil {
			return nil, err
		}
	}

	return t, nil
}

func ParseValidateForm[T any](r *http.Request) (out *T, err error) {
	return ParseValidate[T](httpx.ParseForm, r)
}

func ParseValidateJsonBody[T any](r *http.Request) (out *T, err error) {
	return ParseValidate[T](httpx.ParseJsonBody, r)
}
