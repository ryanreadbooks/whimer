package infra

import (
	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	commentv1 "github.com/ryanreadbooks/whimer/comment/api/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
)

var (
	// 笔记服务
	commenter commentv1.CommentServiceClient
)

func InitCommenter(c *config.Config) {
	commenter = xgrpc.NewRecoverableClient(c.Backend.Comment,
		commentv1.NewCommentServiceClient,
		func(cc commentv1.CommentServiceClient) { commenter = cc })
}

func Commenter() commentv1.CommentServiceClient {
	return commenter
}
