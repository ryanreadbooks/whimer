package service

import (
	"context"

	"github.com/ryanreadbooks/whimer/conductor/internal/biz"
	"github.com/ryanreadbooks/whimer/conductor/internal/config"
)

type Service struct {
	NamespaceService *NamespaceService
	TaskService      *TaskService
	WorkerService    *WorkerService
	ScanService      *ScanService
}

func NewService(c *config.Config, bizz *biz.Biz) *Service {
	return &Service{
		NamespaceService: NewNamespaceService(bizz),
		TaskService:      NewTaskService(bizz),
		WorkerService:    NewWorkerService(bizz),
		ScanService:      NewScanService(c, bizz),
	}
}

func (s *Service) Start(ctx context.Context) {
	s.ScanService.Start(ctx)
}

func (s *Service) Stop() {
	s.ScanService.Stop()
}
