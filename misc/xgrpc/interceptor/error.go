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
	if err == nil {
		return resp, err
	}

	err = xerror.Cause(err)
	// 转成Error对象透传到下游
	xerr, ok := err.(*xerror.Error)
	if ok {
		code = xerror.GrpcCodeFromHttpStatus(xerr.StatusCode)
		msg = xerr.Json()
	} else {
		// 看是否是原生的grpc err
		st, ok := status.FromError(rawErr)
		if ok {
			code = st.Code()
			msg = xerror.ErrOther.Msg(st.Message()).Json()
		} else {
			code = codes.Internal
			msg = xerror.ErrInternal.Msg(rawErr.Error()).Json()
		}
	}

	return resp, status.Error(code, msg)
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

func UnaryClientErrorHandler(ctx context.Context,
	method string,
	req, reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption) error {

	err := invoker(ctx, method, req, reply, cc, opts...)
	if err != nil {
		// err转成Error类型对象
		if st, ok := status.FromError(err); ok {
			var xerr *xerror.Error = xerror.FromJson(st.Message())
			return xerr
		} else {
			return xerror.NewError(xerror.ErrInternal.StatusCode, xerror.ErrInternal.Code, err.Error())
		}
	}

	return err
}
