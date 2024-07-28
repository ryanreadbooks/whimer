package svc

import (
	"github.com/ryanreadbooks/whimer/misc/oss/keygen"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/external"
	"github.com/ryanreadbooks/whimer/note/internal/repo"
)

type ServiceContext struct {
	Config *config.Config

	// utilities
	OssKeyGen *keygen.Generator

	// other service
	NoteSvc *NoteSvc
}

// 初始化一个service
func NewServiceContext(c *config.Config) *ServiceContext {
	dao := repo.New(c)
	ctx := &ServiceContext{
		Config: c,
	}

	// 外部依赖客户端初始化
	external.Init(c)

	// utilities
	ctx.OssKeyGen = keygen.NewGenerator(
		keygen.WithBucket(c.Oss.Bucket),
		keygen.WithPrefix(c.Oss.Prefix),
		keygen.WithPrependBucket(true),
	)

	// 各个子service初始化
	ctx.NoteSvc = NewNoteSvc(ctx, dao)

	return ctx
}
