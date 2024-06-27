package errorx

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"
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
