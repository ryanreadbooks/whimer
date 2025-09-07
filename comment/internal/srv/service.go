package srv

import (
	"github.com/ryanreadbooks/whimer/comment/internal/biz"
	"github.com/ryanreadbooks/whimer/comment/internal/config"
)

type Service struct {
	CommentSrv *CommentSrv
}

// 初始化一个service
func NewService(c *config.Config) *Service {
	s := &Service{}

	// 基础设施初始化
	biz := biz.New()
	s.CommentSrv = NewCommentSrv(s, biz)

	return s
}
