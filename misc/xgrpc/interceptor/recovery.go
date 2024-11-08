package interceptor

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"google.golang.org/grpc"
)

func UnaryServerRecovery(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {

	defer func() {
		if e := recover(); e != nil {
			err = xerror.Wrap(xerror.ErrInternalPanic)
			return
		}
	}()

	return handler(ctx, req)
}
