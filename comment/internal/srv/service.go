package srv

import (
	"github.com/ryanreadbooks/whimer/comment/internal/biz"
	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/comment/internal/infra"
)

type Service struct {
	CommentSrv *CommentSrv
}

// 初始化一个service
func NewService(c *config.Config) *Service {
	s := &Service{}

	// 基础设施初始化
	infra.Init(c)
	biz := biz.New()
	s.CommentSrv = NewCommentSrv(s, biz)

	return s
}
