package worker

import (
	"context"
	"errors"
)

// ErrTaskAborted 任务被终止的错误
var ErrTaskAborted = errors.New("task aborted by conductor")

// Result 任务执行结果
type Result struct {
	// 输出数据（可选）
	Output any

	// 错误信息（失败时设置）
	Error error

	// 失败时是否可重试（默认 false）
	// 只有可重试的错误才会触发 conductor 的重试机制
	Retryable bool
}

// Handler 任务处理函数（简单模式）
// 返回 Result 表示任务结果
type Handler func(ctx context.Context, task *Task) Result

// TaskHandler 任务处理函数（带任务上下文）
// 通过 TaskContext 可以上报进度和监听终止信号
type TaskHandler func(tc TaskContext) Result
