package task

import "github.com/ryanreadbooks/whimer/conductor/internal/biz/model"

const (
	TaskStateInited       = string(model.TaskStateInited)
	TaskStatePendingRetry = string(model.TaskStatePendingRetry)
	TaskStateDispatched   = string(model.TaskStateDispatched)
	TaskStateRunning      = string(model.TaskStateRunning)
	TaskStateSuccess      = string(model.TaskStateSuccess)
	TaskStateFailure      = string(model.TaskStateFailure)
	TaskStateAborted      = string(model.TaskStateAborted)
	TaskStateExpired      = string(model.TaskStateExpired)
)
