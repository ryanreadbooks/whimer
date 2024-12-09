package grpc

import commentv1 "github.com/ryanreadbooks/whimer/comment/sdk/v1"

// 定义不需要检查uid的方法
var uidCheckIgnoredMethods = []string{
	commentv1.ReplyService_PageGetReply_FullMethodName,
	commentv1.ReplyService_PageGetSubReply_FullMethodName,
	commentv1.ReplyService_PageGetDetailedReply_FullMethodName,
	commentv1.ReplyService_GetPinnedReply_FullMethodName,
	commentv1.ReplyService_CountReply_FullMethodName,
	commentv1.ReplyService_BatchCountReply_FullMethodName,
	commentv1.ReplyService_GetReplyLikeCount_FullMethodName,
	commentv1.ReplyService_GetReplyDislikeCount_FullMethodName,
	commentv1.ReplyService_CheckUserOnObject_FullMethodName,
	commentv1.ReplyService_BatchCheckUserOnObject_FullMethodName,
}
