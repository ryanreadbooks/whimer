package interceptor

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"

	"github.com/bufbuild/protovalidate-go"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var (
	// 全局validator
	validator *protovalidate.Validator
)

func init() {
	var err error
	validator, err = protovalidate.New(protovalidate.WithFailFast(true))
	if err != nil {
		panic(err)
	}
}

// 验证请求
// 服务端拦截请求 进行基本的req检验和validate
func UnaryServerValidateHandle(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {

	if req == nil {
		return nil, xerror.ErrNilArg
	}

	if msg, ok := req.(protoreflect.ProtoMessage); ok {
		if err := validator.Validate(msg); err != nil {
			return nil, xerror.ErrArgs.Msg(err.Error())
		}
	}

	return handler(ctx, req)
}
