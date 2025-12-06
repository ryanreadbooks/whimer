package service

import (
	"github.com/ryanreadbooks/whimer/conductor/internal/biz"
	"github.com/ryanreadbooks/whimer/conductor/internal/config"
)

type Service struct {
	TaskService   *TaskService
	WorkerService *WorkerService
}

func NewService(c *config.Config, bizz *biz.Biz) *Service {
	return &Service{
		TaskService:   NewTaskService(bizz.NamespaceBiz, bizz.TaskBiz),
		WorkerService: NewWorkerService(),
	}
}
