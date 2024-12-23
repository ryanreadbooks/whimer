package comment

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	commentv1 "github.com/ryanreadbooks/whimer/comment/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
)

var (
	// 笔记服务
	commenter commentv1.ReplyServiceClient
)

func Init(c *config.Config) {
	commenter = xgrpc.NewRecoverableClient(c.Backend.Comment,
		commentv1.NewReplyServiceClient,
		func(cc commentv1.ReplyServiceClient) { commenter = cc })
}

func Commenter() commentv1.ReplyServiceClient {
	return commenter
}
