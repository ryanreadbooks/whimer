package interceptor

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/stacktrace"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 处理grpc server返回的interceptor
func UnaryServerErrorHandler(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {

	var (
		code   codes.Code = codes.Unknown
		msg    string
		rawErr error
	)
	// 自动输出日志
	defer func() {
		if rawErr != nil {
			errPxy, ok := rawErr.(xerror.ErrProxy)
			if ok && errPxy != nil {
				msg, underErr := xerror.UnwrapMsg(errPxy)
				prepare := xlog.Msg(msg).
					Err(underErr).
					FieldMap(errPxy.Fields()).
					ExtraMap(errPxy.Extra()).
					Extras("stack", stacktrace.FormatFrames(xerror.UnwrapFrames(errPxy)))
				if codeShouldLogError(code) {
					prepare.Errorx(errPxy.Context())
				} else {
					prepare.Infox(errPxy.Context())
				}
			} else {
				prepare := xlog.Msg(rawErr.Error()).Err(rawErr)
				if codeShouldLogError(code) {
					prepare.Errorx(ctx)
				} else {
					prepare.Infox(ctx)
				}
			}
		}
	}()

	resp, err = handler(ctx, req)
	rawErr = err
	if err != nil {
		// 错误转换 自动转换成grpc error
		st, ok := status.FromError(err) // 已经是grpc error
		if ok {
			code = st.Code()
			return resp, err // err == nil is handled here
		}

		err = xerror.Cause(err)
		errx, ok := err.(*xerror.Error)
		if ok {
			code = xerror.GrpcCodeFromHttpStatus(errx.StatusCode)
			msg = errx.Message
		}

		return resp, status.Error(code, msg)
	}

	return resp, err
}

func codeShouldLogError(code codes.Code) bool {
	switch code {
	case codes.OK,
		codes.NotFound,
		codes.InvalidArgument,
		codes.AlreadyExists,
		codes.PermissionDenied,
		codes.Unauthenticated:
		return false
	}
	return true
}
