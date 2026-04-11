package dep

import (
	commentv1 "github.com/ryanreadbooks/whimer/idl/gen/go/comment/api/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/pilot/internal/config"
)

var (
	// 笔记服务
	commenter commentv1.CommentServiceClient
)

func InitCommenter(c *config.Config) {
	commenter = commentv1.NewCommentServiceClient(
		xgrpc.NewRecoverableClientConn(c.Backend.Comment),
	)
}

func Commenter() commentv1.CommentServiceClient {
	return commenter
}
