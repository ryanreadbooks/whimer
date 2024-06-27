package svc

import (
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/repo"
)

type ServiceContext struct {
	Config *config.Config
	Manage *Manage
}

// 初始化一个service
func NewServiceContext(c *config.Config) *ServiceContext {
	dao := repo.New(c)
	return &ServiceContext{
		Config: c,
		Manage: NewManage(dao),
	}
}
