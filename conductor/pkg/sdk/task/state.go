package task

import "github.com/ryanreadbooks/whimer/conductor/internal/biz/model"

// 任务执行状态
const (
	// 初始化完成
	TaskStateInited       = string(model.TaskStateInited)
	
	// 等待重试
	TaskStatePendingRetry = string(model.TaskStatePendingRetry)

	// 任务已下发
	TaskStateDispatched   = string(model.TaskStateDispatched)

	// 任务下发后执行中
	TaskStateRunning      = string(model.TaskStateRunning)

	// 任务执行成功
	TaskStateSuccess      = string(model.TaskStateSuccess)

	// 任务执行失败
	TaskStateFailure      = string(model.TaskStateFailure)

	// 任务被主动终止
	TaskStateAborted      = string(model.TaskStateAborted)

	// 任务执行超时
	TaskStateExpired      = string(model.TaskStateExpired)
)
