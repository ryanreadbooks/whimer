package service

import (
	"context"
	"sync"
	"time"

	"github.com/ryanreadbooks/whimer/conductor/internal/biz"
	"github.com/ryanreadbooks/whimer/conductor/internal/biz/model"
	"github.com/ryanreadbooks/whimer/conductor/internal/biz/shard"
	"github.com/ryanreadbooks/whimer/conductor/internal/config"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

const (
	defaultScanLimit    = 100
	defaultTaskChBuffer = 10000
)

// scanState 扫描状态
type scanState struct {
	mu                 sync.RWMutex
	initedOffset       uuid.UUID
	pendingRetryOffset uuid.UUID
	failureOffset      uuid.UUID
}

func newScanState() *scanState {
	return &scanState{
		initedOffset:       uuid.EmptyUUID(),
		pendingRetryOffset: uuid.EmptyUUID(),
		failureOffset:      uuid.EmptyUUID(),
	}
}

func (s *scanState) GetInitedOffset() uuid.UUID {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.initedOffset
}

func (s *scanState) SetInitedOffset(offset uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.initedOffset = offset
}

func (s *scanState) ResetInited() {
	s.SetInitedOffset(uuid.EmptyUUID())
}

func (s *scanState) GetPendingRetryOffset() uuid.UUID {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.pendingRetryOffset
}

func (s *scanState) SetPendingRetryOffset(offset uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingRetryOffset = offset
}

func (s *scanState) ResetPendingRetry() {
	s.SetPendingRetryOffset(uuid.EmptyUUID())
}

func (s *scanState) GetFailureOffset() uuid.UUID {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.failureOffset
}

func (s *scanState) SetFailureOffset(offset uuid.UUID) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failureOffset = offset
}

func (s *scanState) ResetFailure() {
	s.SetFailureOffset(uuid.EmptyUUID())
}

type ScanService struct {
	conf      *config.Config
	bizz      *biz.Biz
	shardBiz  *shard.Biz
	taskBiz   *biz.TaskBiz
	workerBiz *biz.WorkerBiz

	state *scanState

	// 任务传递 channel
	taskCh chan *model.Task

	quitCh chan struct{}
	doneCh chan struct{}
}

func NewScanService(conf *config.Config, bizz *biz.Biz) *ScanService {
	return &ScanService{
		conf:      conf,
		bizz:      bizz,
		shardBiz:  bizz.ShardBiz,
		taskBiz:   bizz.TaskBiz,
		workerBiz: bizz.WorkerBiz,
		state:     newScanState(),
		taskCh:    make(chan *model.Task, defaultTaskChBuffer),
		quitCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
}

func (s *ScanService) Start(ctx context.Context) {
	// 任务扫描协程（inited + pending_retry）
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "conductor.scan.dispatch",
		Job: func(ctx context.Context) error {
			defer close(s.doneCh)
			s.dispatchScanLoop(ctx)
			return nil
		},
	})

	// 任务分发处理协程
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "conductor.scan.processor",
		Job: func(ctx context.Context) error {
			s.processLoop(ctx)
			return nil
		},
	})

	// 失败任务重试协程
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "conductor.scan.retry",
		Job: func(ctx context.Context) error {
			s.retryScanLoop(ctx)
			return nil
		},
	})

	// 过期任务清理协程
	concurrent.SafeGo2(ctx, concurrent.SafeGo2Opt{
		Name: "conductor.scan.expire",
		Job: func(ctx context.Context) error {
			s.expireScanLoop(ctx)
			return nil
		},
	})
}

func (s *ScanService) Stop() {
	close(s.quitCh)
	<-s.doneCh
}

// ========== 任务分发扫描 ==========

func (s *ScanService) dispatchScanLoop(ctx context.Context) {
	ticker := time.NewTicker(s.conf.ScanConfig.GetProcessInterval())
	defer ticker.Stop()

	for {
		select {
		case <-s.quitCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.doDispatchScan(ctx)
		}
	}
}

func (s *ScanService) doDispatchScan(ctx context.Context) {
	if !s.shardBiz.HasShard() {
		return
	}

	shardRange := s.shardBiz.GetShardRange()

	// 扫描 inited 状态任务
	s.scanInitedTasks(ctx, shardRange)

	// 扫描 pending_retry 状态任务
	s.scanPendingRetryTasks(ctx, shardRange)
}

func (s *ScanService) scanInitedTasks(ctx context.Context, shardRange shard.Range) {
	offset := s.state.GetInitedOffset()

	tasks, err := s.taskBiz.GetInitedTasks(ctx,
		shardRange.Start, shardRange.End,
		defaultScanLimit, offset)
	if err != nil {
		xlog.Msg("scan inited tasks failed").
			Extras("shardRange", shardRange.String(), "offset", offset.String()).
			Err(err).
			Errorx(ctx)
		return
	}

	if len(tasks) == 0 {
		s.state.ResetInited()
		return
	}

	for _, task := range tasks {
		select {
		case <-s.quitCh:
			return
		case <-ctx.Done():
			return
		case s.taskCh <- task:
		}
	}

	lastTask := tasks[len(tasks)-1]
	s.state.SetInitedOffset(lastTask.Id)

	xlog.Msg("scan inited batch completed").
		Extras("shardRange", shardRange.String(),
			"count", len(tasks),
			"newOffset", lastTask.Id.String()).
		Debugx(ctx)
}

func (s *ScanService) scanPendingRetryTasks(ctx context.Context, shardRange shard.Range) {
	offset := s.state.GetPendingRetryOffset()

	tasks, err := s.taskBiz.GetPendingRetryTasks(ctx,
		shardRange.Start, shardRange.End,
		defaultScanLimit, offset)
	if err != nil {
		xlog.Msg("scan pending_retry tasks failed").
			Extras("shardRange", shardRange.String(), "offset", offset.String()).
			Err(err).
			Errorx(ctx)
		return
	}

	if len(tasks) == 0 {
		s.state.ResetPendingRetry()
		return
	}

	for _, task := range tasks {
		select {
		case <-s.quitCh:
			return
		case <-ctx.Done():
			return
		case s.taskCh <- task:
		}
	}

	lastTask := tasks[len(tasks)-1]
	s.state.SetPendingRetryOffset(lastTask.Id)

	xlog.Msg("scan pending_retry batch completed").
		Extras("shardRange", shardRange.String(),
			"count", len(tasks),
			"newOffset", lastTask.Id.String()).
		Debugx(ctx)
}

func (s *ScanService) processLoop(ctx context.Context) {
	for {
		select {
		case <-s.quitCh:
			return
		case <-ctx.Done():
			return
		case task := <-s.taskCh:
			s.processTask(ctx, task)
		}
	}
}

func (s *ScanService) processTask(ctx context.Context, task *model.Task) {
	// 尝试分发任务给等待的 Worker
	dispatched := s.workerBiz.DispatchTask(ctx, task)
	if !dispatched {
		// 没有可用的 Worker，任务留在当前状态，下次扫描会再次处理
		xlog.Msg("no available worker for task").
			Extras("taskId", task.Id.String(), "taskType", task.TaskType).
			Debugx(ctx)
		return
	}

	// 分发成功，更新任务状态为 dispatched
	err := s.bizz.Tx(ctx, func(ctx context.Context) error {
		return s.taskBiz.UpdateTaskState(ctx, task.Id, model.TaskStateDispatched)
	})
	if err != nil {
		xlog.Msg("update task state to dispatched failed").
			Extras("taskId", task.Id.String()).
			Err(err).
			Errorx(ctx)
		return
	}

	xlog.Msg("task dispatched").
		Extras("taskId", task.Id.String(),
			"taskType", task.TaskType,
			"state", string(task.State)).
		Infox(ctx)
}

// ========== 失败任务重试扫描 ==========

func (s *ScanService) retryScanLoop(ctx context.Context) {
	ticker := time.NewTicker(s.conf.ScanConfig.GetRetryInterval())
	defer ticker.Stop()

	for {
		select {
		case <-s.quitCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.doRetryScan(ctx)
		}
	}
}

func (s *ScanService) doRetryScan(ctx context.Context) {
	if !s.shardBiz.HasShard() {
		return
	}

	shardRange := s.shardBiz.GetShardRange()
	offset := s.state.GetFailureOffset()

	tasks, err := s.taskBiz.GetFailureTasks(ctx,
		shardRange.Start, shardRange.End,
		defaultScanLimit, offset)
	if err != nil {
		xlog.Msg("scan failure tasks failed").
			Extras("shardRange", shardRange.String(), "offset", offset.String()).
			Err(err).
			Errorx(ctx)
		return
	}

	if len(tasks) == 0 {
		s.state.ResetFailure()
		return
	}

	for _, task := range tasks {
		s.processFailureTask(ctx, task)
	}

	lastTask := tasks[len(tasks)-1]
	s.state.SetFailureOffset(lastTask.Id)

	xlog.Msg("scan retry batch completed").
		Extras("shardRange", shardRange.String(),
			"count", len(tasks),
			"newOffset", lastTask.Id.String()).
		Debugx(ctx)
}

func (s *ScanService) processFailureTask(ctx context.Context, task *model.Task) {
	// 检查是否已过期
	now := time.Now().UnixMilli()
	if task.IsExpired(now) {
		err := s.bizz.Tx(ctx, func(ctx context.Context) error {
			return s.taskBiz.ExpireTask(ctx, task.Id)
		})
		if err != nil {
			xlog.Msg("expire task failed").
				Extras("taskId", task.Id.String()).
				Err(err).
				Errorx(ctx)
		} else {
			xlog.Msg("task expired, marked as expired").
				Extras("taskId", task.Id.String()).
				Infox(ctx)
		}
		return
	}

	// 获取当前重试次数
	currentRetryCnt, err := s.taskBiz.GetTaskRetryCnt(ctx, task.Id)
	if err != nil {
		xlog.Msg("get task retry cnt failed").
			Extras("taskId", task.Id.String()).
			Err(err).
			Errorx(ctx)
		return
	}

	// 检查是否可以重试
	if !task.CanRetry(int64(currentRetryCnt)) {
		xlog.Msg("task retry limit reached, keeping failure state").
			Extras("taskId", task.Id.String(),
				"maxRetryCnt", task.MaxRetryCnt,
				"currentRetryCnt", currentRetryCnt).
			Infox(ctx)
		return
	}

	// 执行重试：将状态改为 pending_retry
	newRetryCnt := currentRetryCnt + 1
	err = s.bizz.Tx(ctx, func(ctx context.Context) error {
		return s.taskBiz.RetryTask(ctx, task.Id, newRetryCnt)
	})
	if err != nil {
		xlog.Msg("retry task failed").
			Extras("taskId", task.Id.String()).
			Err(err).
			Errorx(ctx)
		return
	}

	xlog.Msg("task marked for retry").
		Extras("taskId", task.Id.String(),
			"retryCnt", newRetryCnt,
			"maxRetryCnt", task.MaxRetryCnt).
		Infox(ctx)
}

// ========== 过期任务清理扫描 ==========

func (s *ScanService) expireScanLoop(ctx context.Context) {
	ticker := time.NewTicker(s.conf.ScanConfig.GetExpireInterval())
	defer ticker.Stop()

	for {
		select {
		case <-s.quitCh:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.doExpireScan(ctx)
		}
	}
}

func (s *ScanService) doExpireScan(ctx context.Context) {
	if !s.shardBiz.HasShard() {
		return
	}

	shardRange := s.shardBiz.GetShardRange()

	tasks, err := s.taskBiz.GetExpiredTasks(ctx,
		shardRange.Start, shardRange.End,
		defaultScanLimit)
	if err != nil {
		xlog.Msg("scan expired tasks failed").
			Extras("shardRange", shardRange.String()).
			Err(err).
			Errorx(ctx)
		return
	}

	if len(tasks) == 0 {
		return
	}

	for _, task := range tasks {
		err := s.bizz.Tx(ctx, func(ctx context.Context) error {
			return s.taskBiz.ExpireTask(ctx, task.Id)
		})
		if err != nil {
			xlog.Msg("expire task failed").
				Extras("taskId", task.Id.String()).
				Err(err).
				Errorx(ctx)
			continue
		}

		xlog.Msg("task expired due to timeout").
			Extras("taskId", task.Id.String(),
				"expireTime", task.ExpireTime).
			Infox(ctx)
	}

	xlog.Msg("scan expire batch completed").
		Extras("shardRange", shardRange.String(),
			"count", len(tasks)).
		Debugx(ctx)
}
