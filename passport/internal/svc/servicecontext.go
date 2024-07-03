package svc

import (
	"github.com/ryanreadbooks/whimer/passport/internal/config"
	"github.com/ryanreadbooks/whimer/passport/internal/repo"
	"github.com/ryanreadbooks/whimer/passport/internal/svc/login"
)

type ServiceContext struct {
	Config *config.Config

	LoginSvc *login.Service
}

// 初始化一个service
func NewServiceContext(c *config.Config) *ServiceContext {
	repo := repo.New(c)
	ctx := &ServiceContext{
		Config:   c,
		LoginSvc: login.New(c, repo),
	}

	return ctx
}
