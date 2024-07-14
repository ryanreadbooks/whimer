package errorx

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/status"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

func init() {
	httpx.SetErrorHandler(ErrorHandler)
	httpx.SetErrorHandlerCtx(ErrorHandlerCtx)
	httpx.SetOkHandler(ResultHandler)
}

func errorHandler(err error) (int, any) {
	if err == nil {
		return http.StatusOK, ErrSuccess
	}

	xerr, ok := err.(*Error)
	if ok {
		return xerr.StatusCode, xerr
	}

	// 一并处理grpc错误
	gerr, ok := status.FromError(err)
	if ok {
		httpCode := runtime.HTTPStatusFromCode(gerr.Code())
		return httpCode, NewError(httpCode, CommonCodeOther, gerr.Message())
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
