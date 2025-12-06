package service

import (
	"context"

	"github.com/ryanreadbooks/whimer/conductor/internal/biz"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

type TaskService struct {
	namespaceBiz *biz.NamespaceBiz
	taskBiz      *biz.TaskBiz
}

func NewTaskService(namespaceBiz *biz.NamespaceBiz, taskBiz *biz.TaskBiz) *TaskService {
	return &TaskService{
		namespaceBiz: namespaceBiz,
		taskBiz:      taskBiz,
	}
}

type RegisterTaskRequest struct {
	TaskType  string `json:"task_type"`
	Namespace string `json:"namespace"`
}

func (s *TaskService) RegisterTask(ctx context.Context,
	req *RegisterTaskRequest) error {

	_, err := s.namespaceBiz.Get(ctx, req.Namespace)
	if err != nil {
		return xerror.Wrapf(err, "task service register task failed").WithExtra("namespace", req.Namespace).WithCtx(ctx)
	}

	return nil
}
