package svc

import (
	"github.com/ryanreadbooks/whimer/passport/internal/config"
	"github.com/ryanreadbooks/whimer/passport/internal/repo"
	"github.com/ryanreadbooks/whimer/passport/internal/svc/signinup"
)

type ServiceContext struct {
	Config *config.Config

	SignInUpSvc *signinup.Service
}

// 初始化一个service
func NewServiceContext(c *config.Config) *ServiceContext {
	repo := repo.New(c)
	ctx := &ServiceContext{
		Config:   c,
		SignInUpSvc: signinup.New(c, repo),
	}

	return ctx
}
