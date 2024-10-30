package xerror

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/status"
)

func init() {
	httpx.SetErrorHandler(ErrorHandler)
	httpx.SetErrorHandlerCtx(ErrorHandlerCtx)
	httpx.SetOkHandler(ResultHandler)
}

func errorHandler(err error) (stcode int, retErr any) {
	if err == nil {
		return http.StatusOK, Success
	}

	var errPxy ErrProxy

	// HTTP接口全局错误日志打印
	defer func() {
		if stcode >= http.StatusInternalServerError {
			if errPxy != nil {
				xlog.Msg(UnwindMsg(errPxy)).
					Err(errPxy).
					FieldMap(errPxy.Fields()).
					ExtraMap(errPxy.Extra()).
					Errorx(errPxy.Context())
			} else {

			}

			return
		}
	}()

	// 全局错误处理
	errPxy, _ = err.(ErrProxy)

	err = Cause(err)
	xerr, ok := err.(*Error)
	if ok {
		return xerr.StatusCode, xerr
	}

	// 一并处理grpc错误
	gerr, ok := status.FromError(err)
	if ok {
		httpCode := runtime.HTTPStatusFromCode(gerr.Code())
		return httpCode, NewError(httpCode, CodeOther, gerr.Message())
	}

	return http.StatusInternalServerError, err
}

func ErrorHandler(err error) (int, any) {
	return errorHandler(err)
}

func ErrorHandlerCtx(ctx context.Context, err error) (int, any) {
	return errorHandler(err)
}

func ResultHandler(ctx context.Context, data any) any {
	return NewResult("success", data)
}
