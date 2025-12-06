package biz

import (
	"context"

	"github.com/ryanreadbooks/whimer/conductor/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/misc/uuid"
)

type TaskBiz struct {
	taskDao        *dao.TaskDao
	taskHistoryDao *dao.TaskHistoryDao
	lockDao        *dao.LockDao
}

func NewTaskBiz(
	taskDao *dao.TaskDao,
	taskHistoryDao *dao.TaskHistoryDao,
	lockDao *dao.LockDao,
) *TaskBiz {
	return &TaskBiz{
		taskDao:        taskDao,
		taskHistoryDao: taskHistoryDao,
		lockDao:        lockDao,
	}
}

type RegisterTaskRequest struct {
	TaskType  string `json:"task_type"`
	Namespace string `json:"namespace"`
}

type RegisterTaskResponse struct {
	TaskId string `json:"task_id"`
}

// 创建任务
func (b *TaskBiz) RegisterTask(ctx context.Context, req *RegisterTaskRequest) (*RegisterTaskResponse, error) {
	// 检查taskType

	// 入库即完成注册
	taskId := uuid.NewUUID()

	return &RegisterTaskResponse{
		TaskId: taskId.String(),
	}, nil
}

// 终止任务
func (b *TaskBiz) AbortTask(ctx context.Context, taskId string) error {

	return nil
}
