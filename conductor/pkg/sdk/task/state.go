package task

import "github.com/ryanreadbooks/whimer/conductor/internal/biz/model"

// 任务执行状态
const (
	// 任务下发后执行中
	TaskStateRunning = string(model.TaskStateRunning)

	// 任务执行成功
	TaskStateSuccess = string(model.TaskStateSuccess)

	// 任务执行失败
	TaskStateFailure = string(model.TaskStateFailure)
)
