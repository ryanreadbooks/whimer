package xerror

type Result struct {
	Code int    `json:"code"` // 业务响应状态码
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func NewResult(msg string, data any) *Result {
	return &Result{
		Code: 0, // 固定为0表示成功
		Msg:  msg,
		Data: data,
	}
}
