package model

import (
	"encoding/json"

	"github.com/ryanreadbooks/whimer/conductor/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

type TaskState string

const (
	TaskStateInited       TaskState = "inited"        // 已创建（首次）
	TaskStatePendingRetry TaskState = "pending_retry" // 待重试（失败后等待重新分发）
	TaskStateDispatched   TaskState = "dispatched"    // 已分发给 Worker
	TaskStateRunning      TaskState = "running"       // Worker 正在执行
	TaskStateSuccess      TaskState = "success"       // 执行成功
	TaskStateFailure      TaskState = "failure"       // 执行失败（最终失败，不再重试）
	TaskStateAborted      TaskState = "aborted"       // 已取消
	TaskStateExpired      TaskState = "expired"       // 已过期
)

// IsPending 是否处于待分发状态（inited 或 pending_retry）
func (s TaskState) IsPending() bool {
	return s == TaskStateInited || s == TaskStatePendingRetry
}

// IsTerminal 是否为终态（success, failure, aborted, expired）
func (s TaskState) IsTerminal() bool {
	return s == TaskStateSuccess || s == TaskStateFailure || s == TaskStateAborted || s == TaskStateExpired
}

// ExternalState 对外展示的状态，只有三种：success, failure, running
// 内部的 inited, pending_retry, dispatched, running, aborted, expired 都归为 running
func (s TaskState) ExternalState() TaskState {
	switch s {
	case TaskStateSuccess:
		return TaskStateSuccess
	case TaskStateFailure:
		return TaskStateFailure
	default:
		return TaskStateRunning
	}
}

type Task struct {
	Id          uuid.UUID
	Namespace   string
	TaskType    string
	InputArgs   []byte
	OutputArgs  []byte
	CallbackUrl string
	State       TaskState
	TraceId     string
	MaxRetryCnt int64        // -1 无限重试, 0 不重试
	CurRetryCnt int64        // 当前已重试次数
	ExpireTime  int64        // 过期时间 unix ms
	Settings    TaskSettings // 额外设置
	Ctime       int64
	Utime       int64
}

// TaskSettings 任务额外设置（可选）
type TaskSettings struct {
	// 可扩展的额外配置
	Extra map[string]any `json:"extra,omitempty"`
}

func TaskFromPO(po *dao.TaskPO) *Task {
	var settings TaskSettings
	if len(po.Settings) > 0 {
		err := json.Unmarshal(po.Settings, &settings)
		if err != nil {
			xlog.Msgf("failed to unmarshal task settings: %v", err).Err(err).Error()
		}
	}

	return &Task{
		Id:          po.Id,
		Namespace:   po.Namespace,
		TaskType:    po.TaskType,
		InputArgs:   po.InputArgs,
		OutputArgs:  po.OutputArgs,
		CallbackUrl: po.CallbackUrl,
		State:       TaskState(po.State),
		TraceId:     po.TraceId,
		MaxRetryCnt: po.MaxRetryCnt,
		CurRetryCnt: po.CurRetryCnt,
		ExpireTime:  po.ExpireTime,
		Settings:    settings,
		Ctime:       po.Ctime,
		Utime:       po.Utime,
	}
}

// IsExpired 检查任务是否已过期
func (t *Task) IsExpired(now int64) bool {
	return t.ExpireTime > 0 && now > t.ExpireTime
}

// CanRetry 检查任务是否可以重试
func (t *Task) CanRetry() bool {
	if t.MaxRetryCnt < 0 {
		return true // -1 表示无限重试
	}
	return t.CurRetryCnt < t.MaxRetryCnt
}

type TaskHistory struct {
	TaskId uuid.UUID
	State  TaskState
	Ctime  int64
}

func TaskHistoryFromPO(po *dao.TaskHistoryPO) *TaskHistory {
	return &TaskHistory{
		TaskId: po.TaskId,
		State:  TaskState(po.State),
		Ctime:  po.Ctime,
	}
}
