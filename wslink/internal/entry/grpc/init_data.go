package grpc

import (
	pushv1 "github.com/ryanreadbooks/whimer/idl/gen/go/wslink/api/push/v1"
)

// 定义不需要检查uid的方法
var uidCheckIgnoredMethods = []string{
	pushv1.PushService_Push_FullMethodName,
}
