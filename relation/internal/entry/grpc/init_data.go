package grpc

import relationv1 "github.com/ryanreadbooks/whimer/relation/sdk/v1"

// 定义不需要检查uid的方法
var uidCheckIgnoredMethods = []string{
	relationv1.RelationService_GetUserFanCount_FullMethodName,
	relationv1.RelationService_GetUserFollowingCount_FullMethodName,
}
