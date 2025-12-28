package worker

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
)

// TaskContext 任务执行上下文接口，提供进度上报和终止检测能力
type TaskContext interface {
	// Context 获取可取消的 context，当任务被终止时会自动取消
	Context() context.Context

	// Task 获取任务信息
	Task() *Task

	// IsAborted 检查任务是否已被终止
	IsAborted() bool

	// AbortCh 返回一个 channel，当任务被终止时会关闭
	// 可用于 select 监听终止信号
	AbortCh() <-chan struct{}

	// ReportProgress 主动上报当前任务进度（0-100）
	// 返回 true 表示需要终止任务
	ReportProgress(progress int64) bool

	// SetProgressProvider 设置进度提供者
	// 设置后心跳上报时会自动获取进度
	SetProgressProvider(provider ProgressProvider)
}

// ProgressProvider 进度提供者接口
// 实现此接口后，心跳上报时会自动调用 Progress() 获取当前进度
type ProgressProvider interface {
	Progress() int64 // 返回当前进度 (0-100)，返回 -1 表示不上报具体进度
}

// ProgressFunc 进度函数类型，方便使用函数作为 ProgressProvider
type ProgressFunc func() int64

func (f ProgressFunc) Progress() int64 { return f() }

// taskContextImpl TaskContext 的内部实现
type taskContextImpl struct {
	task             *Task
	worker           *Worker
	ctx              context.Context
	cancel           context.CancelFunc
	aborted          atomic.Bool
	abortCh          chan struct{}
	interval         time.Duration // 上报间隔
	progressProvider atomic.Value  // ProgressProvider
}

// newTaskContext 创建任务执行上下文
func newTaskContext(ctx context.Context, task *Task, worker *Worker, interval time.Duration) *taskContextImpl {
	ctx, cancel := context.WithCancel(ctx)
	return &taskContextImpl{
		task:     task,
		worker:   worker,
		ctx:      ctx,
		cancel:   cancel,
		abortCh:  make(chan struct{}),
		interval: interval,
	}
}

func (tc *taskContextImpl) Context() context.Context {
	return tc.ctx
}

func (tc *taskContextImpl) Task() *Task {
	return tc.task
}

func (tc *taskContextImpl) IsAborted() bool {
	return tc.aborted.Load()
}

func (tc *taskContextImpl) AbortCh() <-chan struct{} {
	return tc.abortCh
}

func (tc *taskContextImpl) ReportProgress(progress int64) bool {
	if tc.aborted.Load() {
		return true
	}

	aborted := tc.worker.reportTaskProgress(tc.ctx, tc.task.Id, progress)
	if aborted {
		tc.markAborted()
	}
	return aborted
}

func (tc *taskContextImpl) SetProgressProvider(provider ProgressProvider) {
	tc.progressProvider.Store(provider)
}

func (tc *taskContextImpl) getProgress() int64 {
	if p := tc.progressProvider.Load(); p != nil {
		return p.(ProgressProvider).Progress()
	}
	return -1 // 未设置 provider 时不上报具体进度
}

// markAborted 标记任务为已终止
func (tc *taskContextImpl) markAborted() {
	if tc.aborted.CompareAndSwap(false, true) {
		close(tc.abortCh)
		tc.cancel()
	}
}

// startHeartbeat 启动心跳上报，定期检测任务是否被终止
func (tc *taskContextImpl) startHeartbeat() {
	concurrent.SafeGo2(tc.ctx, concurrent.SafeGo2Opt{
		Name:             "conductor.worker.heartbeat",
		InheritCtxCancel: true, // 继承 ctx 取消，任务结束时心跳自动退出
		Job: func(ctx context.Context) error {
			ticker := time.NewTicker(tc.interval)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					return nil
				case <-ticker.C:
					if tc.aborted.Load() {
						return nil
					}
					// 心跳上报，通过 ProgressProvider 获取当前进度
					progress := tc.getProgress()
					aborted := tc.worker.reportTaskProgress(ctx, tc.task.Id, progress)
					if aborted {
						tc.markAborted()
						return nil
					}
				}
			}
		},
	})
}

// stop 停止任务上下文
func (tc *taskContextImpl) stop() {
	tc.cancel()
}
