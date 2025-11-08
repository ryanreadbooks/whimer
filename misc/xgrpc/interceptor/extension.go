package interceptor

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xgrpc/extension"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor/checker"
	"google.golang.org/grpc"
)

var (
	// 全局extension
	extensions *extension.Extension
)

func init() {
	extensions = extension.NewExtension()
}

func UnaryServerExtensionHandler(ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp any, err error) {

	extOpt := extensions.Get(info.FullMethod)
	if extOpt != nil && extOpt.SkipMetadataUidCheck {
		// 跳过后续检查
		ctx = checker.ForceSkipUidCheck(ctx)
	}

	return handler(ctx, req)
}
