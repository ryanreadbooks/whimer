package svc

import (
	"github.com/ryanreadbooks/whimer/counter/internal/config"
	"github.com/ryanreadbooks/whimer/counter/internal/repo"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config    *config.Config
	RecordSvc *CounterSvc
}

// 初始化一个service
func NewServiceContext(c *config.Config) *ServiceContext {
	dao := repo.New(c)
	ctx := &ServiceContext{
		Config: c,
	}

	cache := redis.MustNewRedis(c.Redis)

	ctx.RecordSvc = NewCounterSvc(ctx, dao, cache)
	return ctx
}
