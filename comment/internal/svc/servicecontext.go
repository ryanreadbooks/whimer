package svc

import (
	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/comment/internal/external"
	"github.com/ryanreadbooks/whimer/comment/internal/repo"
)

type ServiceContext struct {
	Config *config.Config
	CommentSvc *CommentSvc
}

// 初始化一个service
func NewServiceContext(c *config.Config) *ServiceContext {
	dao := repo.New(c)
	ctx := &ServiceContext{
		Config: c,
	}

	// 外部依赖客户端初始化
	external.Init(c)

	// 各个子service初始化
	ctx.CommentSvc = NewCommentSvc(ctx, dao)

	return ctx
}
