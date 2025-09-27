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
	httpx.SetErrorHandlerCtx(ErrorHandlerCtx)
	httpx.SetOkHandler(ResultHandler)
}

func errorHandler(ctx context.Context, err error) (stcode int, retErr any) {
	if err == nil {
		return http.StatusOK, Success
	}

	// 全局错误处理
	errPxy, _ := err.(ErrProxy)

	// HTTP接口全局错误日志打印
	defer func() {
		if errPxy != nil {
			// msg, underErr := UnwrapMsg(errPxy)
			prepare := xlog.Err(errPxy).
				FieldMap(errPxy.Fields()).
				ExtraMap(errPxy.Extra())
				// 5XX 打印ERROR
			if stcode >= http.StatusInternalServerError {
				prepare.Errorx(errPxy.Context())
			} else {
				// 4xx 打印INFO
				prepare.Infox(errPxy.Context())
			}
		} else {
			prepare := xlog.Msg(err.Error())
			if stcode >= http.StatusInternalServerError {
				prepare.Errorx(ctx)
			} else {
				prepare.Infox(ctx)
			}
		}
	}()

	err = Cause(err)
	xerr, ok := err.(*Error)
	if ok {
		return xerr.StatusCode, xerr.AsResult()
	}

	// 尝试解析处理grpc错误
	gerr, ok := status.FromError(err)
	if ok {
		httpCode := runtime.HTTPStatusFromCode(gerr.Code())
		return httpCode, NewError(httpCode, CodeOther, gerr.Message()).AsResult()
	}

	return http.StatusInternalServerError, err
}

func ErrorHandler(err error) (int, any) {
	return errorHandler(context.Background(), err)
}

func ErrorHandlerCtx(ctx context.Context, err error) (int, any) {
	return errorHandler(ctx, err)
}

func ResultHandler(ctx context.Context, data any) any {
	return NewResult("success", data)
}
