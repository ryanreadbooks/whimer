package svc

import (
	"github.com/ryanreadbooks/whimer/misc/oss/keygen"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/repo"
)

type ServiceContext struct {
	Config *config.Config

	// utilities
	KeyGen *keygen.Generator

	// other service
	Manage *Manage
}

// 初始化一个service
func NewServiceContext(c *config.Config) *ServiceContext {
	dao := repo.New(c)
	ctx := &ServiceContext{
		Config: c,
	}

	// utilities
	ctx.KeyGen = keygen.NewGenerator(
		keygen.WithBucket(c.Oss.Bucket),
		keygen.WithPrefix(c.Oss.Prefix),
		keygen.WithPrependBucket(true),
	)

	// other services
	ctx.Manage = NewManage(dao)
	ctx.Manage.Ctx = ctx

	return ctx
}
