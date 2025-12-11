package biz

import (
	"container/list"
	"context"
	"sync"
	"time"

	"github.com/ryanreadbooks/whimer/conductor/internal/biz/model"
	"github.com/ryanreadbooks/whimer/conductor/internal/config"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

// waitingWorker 等待任务的 Worker
type waitingWorker struct {
	workerId string
	taskType string
	taskCh   chan *model.Task
	doneCh   chan struct{}
	element  *list.Element // 在链表中的位置，用于 O(1) 删除
}

// WorkerBiz 管理 Worker 长轮询
type WorkerBiz struct {
	conf *config.Config

	mu sync.Mutex

	// 按 task_type 分组的等待 Worker 链表 (FIFO)
	waitingWorkers map[string]*list.List
}

func NewWorkerBiz(conf *config.Config) *WorkerBiz {
	return &WorkerBiz{
		conf:           conf,
		waitingWorkers: make(map[string]*list.List),
	}
}

// Worker 等待任务
func (b *WorkerBiz) WaitForTask(
	ctx context.Context,
	workerId, taskType string,
) (*model.Task, error) {
	timeout := b.conf.WorkerConfig.GetLongPollTimeout()
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	w := &waitingWorker{
		workerId: workerId,
		taskType: taskType,
		taskCh:   make(chan *model.Task, 1),
		doneCh:   make(chan struct{}),
	}

	b.addWaiting(w)

	select {
	case task := <-w.taskCh:
		xlog.Msg("worker received task").
			Extras("workerId", workerId,
				"taskType", taskType,
				"taskId", task.Id.String(),
				"traceId", task.TraceId).
			Infox(ctx)
		return task, nil
	case <-timer.C:
		b.removeWaiting(w)
		close(w.doneCh)
		return &model.Task{Id: uuid.EmptyUUID()}, nil
	case <-ctx.Done():
		b.removeWaiting(w)
		close(w.doneCh)
		return nil, ctx.Err()
	}
}

// DispatchTask 分发任务给等待的 Worker
func (b *WorkerBiz) DispatchTask(ctx context.Context, task *model.Task) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.dispatchTaskLocked(ctx, task)
}

func (b *WorkerBiz) dispatchTaskLocked(ctx context.Context, task *model.Task) bool {
	l := b.waitingWorkers[task.TaskType]
	if l == nil || l.Len() == 0 {
		return false
	}

	for {
		front := l.Front()
		if front == nil {
			return false
		}

		w := front.Value.(*waitingWorker)
		l.Remove(front)
		w.element = nil

		// 跳过已超时的 worker
		select {
		case <-w.doneCh:
			continue
		default:
		}

		// 发送任务
		select {
		case w.taskCh <- task:
			xlog.Msg("task dispatched to worker").
				Extras("workerId", w.workerId,
					"taskType", task.TaskType,
					"taskId", task.Id.String()).
				Infox(ctx)
			return true
		default:
			continue
		}
	}
}

// GetWaitingCount 获取指定 taskType 的等待 Worker 数量
func (b *WorkerBiz) GetWaitingCount(taskType string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	if l := b.waitingWorkers[taskType]; l != nil {
		return l.Len()
	}
	return 0
}

// GetTotalWaitingCount 获取所有等待 Worker 数量
func (b *WorkerBiz) GetTotalWaitingCount() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	total := 0
	for _, l := range b.waitingWorkers {
		total += l.Len()
	}
	return total
}

func (b *WorkerBiz) addWaiting(w *waitingWorker) {
	b.mu.Lock()
	defer b.mu.Unlock()

	l := b.waitingWorkers[w.taskType]
	if l == nil {
		l = list.New()
		b.waitingWorkers[w.taskType] = l
	}
	w.element = l.PushBack(w)

	xlog.Msg("worker added to waiting queue").
		Extras("workerId", w.workerId,
			"taskType", w.taskType,
			"queueSize", l.Len()).
		Debug()
}

func (b *WorkerBiz) removeWaiting(w *waitingWorker) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if w.element == nil {
		return
	}

	l := b.waitingWorkers[w.taskType]
	if l != nil {
		l.Remove(w.element)
		w.element = nil
	}
}
