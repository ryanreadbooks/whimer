package svc

import (
	"github.com/ryanreadbooks/whimer/passport/internal/config"
	"github.com/ryanreadbooks/whimer/passport/internal/repo"
	"github.com/ryanreadbooks/whimer/passport/internal/svc/access"
	"github.com/ryanreadbooks/whimer/passport/internal/svc/profile"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config     *config.Config
	AccessSvc  *access.Service
	ProfileSvc *profile.Service
}

// 初始化一个service
func NewServiceContext(c *config.Config) *ServiceContext {
	repo := repo.New(c)
	cache := redis.MustNewRedis(c.Redis)

	ctx := &ServiceContext{
		Config:     c,
		AccessSvc:  access.New(c, repo, cache),
		ProfileSvc: profile.New(c, repo),
	}

	return ctx
}
