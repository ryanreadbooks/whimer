package worker

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	workerv1 "github.com/ryanreadbooks/whimer/conductor/api/worker/v1"
	workerservice "github.com/ryanreadbooks/whimer/conductor/api/workerservice/v1"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xconf"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Options Worker 配置
type Options struct {
	HostConf xconf.Discovery

	// Worker ID（可选）
	WorkerId string

	// Worker IP（可选）
	IP string

	// 并发数（同时处理的任务数）
	Concurrency int

	// 上报重试配置
	ReportRetry RetryOptions

	// 心跳上报间隔（默认 10s）
	// 用于定期检测任务是否被终止
	HeartbeatInterval time.Duration
}

// RetryOptions 重试配置
type RetryOptions struct {
	// 最大重试次数（默认 5）
	MaxAttempts int
	// 初始退避时间（默认 100ms）
	InitialBackoff time.Duration
	// 最大退避时间（默认 10s）
	MaxBackoff time.Duration
	// 退避倍数（默认 2）
	Multiplier float64
}

// Worker 任务执行器
type Worker struct {
	opts         Options
	client       workerservice.WorkerServiceClient
	handlers     map[string]Handler
	taskHandlers map[string]TaskHandler // 带上下文的处理函数
	mu           sync.RWMutex

	quitCh   chan struct{}
	doneCh   chan struct{}
	cancel   context.CancelFunc
	stopping atomic.Bool
}

const defaultHeartbeatInterval = 10 * time.Second

// New 创建 Worker
func New(opts Options) (*Worker, error) {
	if opts.Concurrency <= 0 {
		opts.Concurrency = 1
	}
	if opts.ReportRetry.MaxAttempts <= 0 {
		opts.ReportRetry.MaxAttempts = 5
	}
	if opts.ReportRetry.InitialBackoff <= 0 {
		opts.ReportRetry.InitialBackoff = 100 * time.Millisecond
	}
	if opts.ReportRetry.MaxBackoff <= 0 {
		opts.ReportRetry.MaxBackoff = 10 * time.Second
	}
	if opts.ReportRetry.Multiplier <= 0 {
		opts.ReportRetry.Multiplier = 2
	}
	if opts.HeartbeatInterval <= 0 {
		opts.HeartbeatInterval = defaultHeartbeatInterval
	}

	w := &Worker{
		opts:         opts,
		handlers:     make(map[string]Handler),
		taskHandlers: make(map[string]TaskHandler),
		quitCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
	}

	client := xgrpc.NewRecoverableClient(
		opts.HostConf,
		workerservice.NewWorkerServiceClient,
		func(t workerservice.WorkerServiceClient) {
			w.client = t
		},
		xgrpc.WithoutDefaultInterceptor(),
	)

	w.client = client

	return w, nil
}

// RegisterHandler 注册任务处理函数（简单模式）
func (w *Worker) RegisterHandler(taskType string, handler Handler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers[taskType] = handler
}

// RegisterTaskHandler 注册任务处理函数（带上下文模式）
// 支持进度上报和终止信号检测
func (w *Worker) RegisterTaskHandler(taskType string, handler TaskHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.taskHandlers[taskType] = handler
}

// Run 启动 Worker（阻塞运行）
func (w *Worker) Run(ctx context.Context) error {
	w.mu.RLock()
	taskTypeSet := make(map[string]struct{})
	for taskType := range w.handlers {
		taskTypeSet[taskType] = struct{}{}
	}
	for taskType := range w.taskHandlers {
		taskTypeSet[taskType] = struct{}{}
	}
	taskTypes := make([]string, 0, len(taskTypeSet))
	for taskType := range taskTypeSet {
		taskTypes = append(taskTypes, taskType)
	}
	w.mu.RUnlock()

	if len(taskTypes) == 0 {
		xlog.Msg("no handlers registered, worker exiting").Info()
		return nil
	}

	ctx, w.cancel = context.WithCancel(ctx)

	var wg sync.WaitGroup
	sem := make(chan struct{}, w.opts.Concurrency)

	for _, taskType := range taskTypes {
		wg.Add(1)
		concurrent.SimpleSafeGo(ctx, "conductor.worker.poll_loop", func(ctx context.Context) error {
			defer wg.Done()
			w.pollLoop(ctx, taskType, sem)
			return nil
		})
	}

	wg.Wait()
	close(w.doneCh)
	return nil
}

// Stop 停止 Worker
func (w *Worker) Stop() {
	w.stopping.Store(true)
	close(w.quitCh)
	if w.cancel != nil {
		w.cancel()
	}
	<-w.doneCh
}

func (w *Worker) pollLoop(ctx context.Context, taskType string, sem chan struct{}) {
	for {
		select {
		case <-w.quitCh:
			return
		case <-ctx.Done():
			return
		default:
		}

		select {
		case <-w.quitCh:
			return
		case <-ctx.Done():
			return
		case sem <- struct{}{}:
		}

		task, err := w.poll(ctx, taskType)
		if err != nil {
			if w.stopping.Load() {
				<-sem
				return
			}
			if !isTimeoutError(err) {
				xlog.Msg("poll task failed").Err(err).Errorx(ctx)
				time.Sleep(time.Second)
			}
			<-sem
			continue
		}

		if task == nil || task.Id == "" {
			<-sem
			continue
		}

		concurrent.SimpleSafeGo(ctx, "conductor.worker.process_task", func(ctx context.Context) error {
			defer func() { <-sem }()
			w.processTask(ctx, task)
			return nil
		})
	}
}

func (w *Worker) poll(ctx context.Context, taskType string) (*Task, error) {
	resp, err := w.client.LongPoll(ctx, &workerservice.LongPollRequest{
		Worker: &workerv1.Worker{
			Id: w.opts.WorkerId,
			Ability: &workerv1.WorkerAbility{
				TaskType: taskType,
			},
			Metadata: &workerv1.WorkerMetadata{
				Ip: w.opts.IP,
			},
			State: workerv1.WorkerState_WORKER_STATE_READY,
		},
	})
	if err != nil {
		return nil, err
	}

	return taskFromProto(resp.Task), nil
}

func (w *Worker) processTask(ctx context.Context, task *Task) {
	_, err := w.client.AcceptTask(ctx, &workerservice.AcceptTaskRequest{
		TaskId: task.Id,
	})
	if err != nil {
		xlog.Msg("accept task failed").Extras("taskId", task.Id).Err(err).Errorx(ctx)
		return
	}

	w.mu.RLock()
	handler, hasHandler := w.handlers[task.TaskType]
	taskHandler, hasTaskHandler := w.taskHandlers[task.TaskType]
	w.mu.RUnlock()

	if !hasHandler && !hasTaskHandler {
		xlog.Msg("no handler for task type").Extras("taskType", task.TaskType).Errorx(ctx)
		w.completeTask(ctx, task.Id, nil, false, "no handler for task type", false)
		return
	}

	// AcceptTask 成功后立即启动心跳上报，检测任务是否被终止
	tc := newTaskContext(ctx, task, w, w.opts.HeartbeatInterval)
	tc.startHeartbeat()
	defer tc.stop()

	var result Result

	// 优先使用 TaskHandler（支持进度上报和终止检测）
	if hasTaskHandler {
		result = w.safeExecuteWithContext(tc, taskHandler)
	} else {
		// 简单 Handler 也使用 TaskContext 的 context，支持终止信号
		result = w.safeExecute(tc.Context(), task, handler)
	}

	// 如果任务被终止，不上报结果（服务端已知状态）
	if tc.IsAborted() {
		xlog.Msg("task aborted, skip complete").Extras("taskId", task.Id).Infox(ctx)
		return
	}

	if result.Error != nil {
		w.completeTask(ctx, task.Id, result.Output, false, result.Error.Error(), result.Retryable)
	} else {
		w.completeTask(ctx, task.Id, result.Output, true, "", false)
	}
}

func (w *Worker) safeExecute(ctx context.Context, task *Task, handler Handler) (result Result) {
	defer func() {
		if r := recover(); r != nil {
			xlog.Msg("handler panic").Extras("taskId", task.Id, "panic", r).Errorx(ctx)
			result = Result{Error: &panicError{v: r}}
		}
	}()

	return handler(ctx, task)
}

func (w *Worker) safeExecuteWithContext(tc TaskContext, handler TaskHandler) (result Result) {
	defer func() {
		if r := recover(); r != nil {
			xlog.Msg("handler panic").Extras("taskId", tc.Task().Id, "panic", r).Errorx(tc.Context())
			result = Result{Error: &panicError{v: r}}
		}
	}()

	return handler(tc)
}

// reportTaskProgress 上报任务进度，返回是否需要终止
func (w *Worker) reportTaskProgress(ctx context.Context, taskId string, progress int64) bool {
	resp, err := w.client.ReportTask(ctx, &workerservice.ReportTaskRequest{
		TaskId:   taskId,
		Progress: progress,
	})
	if err != nil {
		// 上报失败时不终止任务，只记录日志
		xlog.Msg("report task progress failed").Extras("taskId", taskId).Err(err).Infox(ctx)
		return false
	}

	return resp.Aborted
}

func (w *Worker) completeTask(
	ctx context.Context,
	taskId string,
	output any,
	success bool,
	errMsg string,
	retryable bool,
) {
	var outputArgs []byte
	if output != nil {
		outputArgs, _ = json.Marshal(output)
	}

	req := &workerservice.CompleteTaskRequest{
		TaskId:     taskId,
		OutputArgs: outputArgs,
		Success:    success,
		ErrorMsg:   []byte(errMsg),
		Retryable:  retryable,
	}

	ctx = context.WithoutCancel(ctx)

	backoff := w.opts.ReportRetry.InitialBackoff
	for attempt := 1; attempt <= w.opts.ReportRetry.MaxAttempts; attempt++ {
		_, err := w.client.CompleteTask(ctx, req)
		if err == nil {
			return
		}

		if attempt == w.opts.ReportRetry.MaxAttempts {
			xlog.Msg("complete task failed after max attempts").
				Extras("taskId", taskId, "attempts", attempt).
				Err(err).Errorx(ctx)
			return
		}

		xlog.Msg("complete task failed, retrying").
			Extras("taskId", taskId, "attempt", attempt, "maxAttempts", w.opts.ReportRetry.MaxAttempts).
			Err(err).Infox(ctx)

		time.Sleep(backoff)
		backoff = min(time.Duration(float64(backoff)*w.opts.ReportRetry.Multiplier), w.opts.ReportRetry.MaxBackoff)
	}
}

type panicError struct {
	v any
}

func (e *panicError) Error() string {
	return "handler panic"
}

func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	if s, ok := status.FromError(err); ok {
		return s.Code() == codes.DeadlineExceeded
	}
	return false
}
