package interceptor

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 处理grpc server返回的interceptor
func UnaryServerErrorHandler(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {

	// 自动输出日志
	defer func() {
		if err != nil {
			// TODO 输出日志
		}
	}()

	resp, err = handler(ctx, req)
	if err != nil {
		// 错误转换 自动转换成grpc error
		resp, err = handler(ctx, req)
		// 处理err
		_, ok := status.FromError(err) // 已经是grpc error
		if ok {
			return resp, err // err == nil is handled here
		}

		var (
			code codes.Code = codes.Unknown
			msg  string     = err.Error()
		)

		errx, ok := err.(*xerror.Error)
		if ok {
			code = xerror.GrpcCodeFromHttpStatus(errx.StatusCode)
			msg = errx.Message
		}

		return resp, status.Error(code, msg)
	}

	return resp, err
}
