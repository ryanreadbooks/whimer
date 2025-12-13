package worker

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	taskv1 "github.com/ryanreadbooks/whimer/conductor/api/task/v1"
	workerv1 "github.com/ryanreadbooks/whimer/conductor/api/worker/v1"
	workerservice "github.com/ryanreadbooks/whimer/conductor/api/workerservice/v1"
	"github.com/ryanreadbooks/whimer/misc/xconf"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Task 任务上下文
type Task struct {
	Id          string
	Namespace   string
	TaskType    string
	InputArgs   []byte
	CallbackUrl string
	MaxRetryCnt int64
	ExpireTime  int64
	Ctime       int64
	TraceId     string
}

// UnmarshalInput 反序列化输入参数
func (t *Task) UnmarshalInput(v any) error {
	if len(t.InputArgs) == 0 {
		return nil
	}
	return json.Unmarshal(t.InputArgs, v)
}

func taskFromProto(t *taskv1.Task) *Task {
	if t == nil {
		return nil
	}
	return &Task{
		Id:          t.Id,
		Namespace:   t.Namespace,
		TaskType:    t.TaskType,
		InputArgs:   t.InputArgs,
		CallbackUrl: t.CallbackUrl,
		MaxRetryCnt: t.MaxRetryCnt,
		ExpireTime:  t.ExpireTime,
		Ctime:       t.Ctime,
		TraceId:     t.TraceId,
	}
}

// Result 任务执行结果
type Result struct {
	// 输出数据（可选）
	Output any

	// 错误信息（失败时设置）
	Error error
}

// Handler 任务处理函数
// 返回 Result 表示任务结果
type Handler func(ctx context.Context, task *Task) Result

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
	opts     Options
	client   workerservice.WorkerServiceClient
	handlers map[string]Handler
	mu       sync.RWMutex

	quitCh   chan struct{}
	doneCh   chan struct{}
	cancel   context.CancelFunc
	stopping atomic.Bool
}

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

	w := &Worker{
		opts:     opts,
		handlers: make(map[string]Handler),
		quitCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
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

// RegisterHandler 注册任务处理函数
func (w *Worker) RegisterHandler(taskType string, handler Handler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.handlers[taskType] = handler
}

// Run 启动 Worker（阻塞运行）
func (w *Worker) Run(ctx context.Context) error {
	w.mu.RLock()
	taskTypes := make([]string, 0, len(w.handlers))
	for taskType := range w.handlers {
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
		go func() {
			defer func() {
				if r := recover(); r != nil {
					xlog.Msg("poll loop panic").Extras("taskType", taskType, "panic", r).Errorx(ctx)
				}
				wg.Done()
			}()
			w.pollLoop(ctx, taskType, sem)
		}()
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

		go func() {
			defer func() {
				if r := recover(); r != nil {
					xlog.Msg("process task panic").Extras("taskId", task.Id, "panic", r).Errorx(ctx)
				}
				<-sem
			}()
			w.processTask(ctx, task)
		}()
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
	handler, ok := w.handlers[task.TaskType]
	w.mu.RUnlock()

	if !ok {
		xlog.Msg("no handler for task type").Extras("taskType", task.TaskType).Errorx(ctx)
		w.completeTask(ctx, task.Id, nil, false, "no handler for task type")
		return
	}

	result := w.safeExecute(ctx, task, handler)

	if result.Error != nil {
		w.completeTask(ctx, task.Id, result.Output, false, result.Error.Error())
	} else {
		w.completeTask(ctx, task.Id, result.Output, true, "")
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

func (w *Worker) completeTask(ctx context.Context, taskId string, output any, success bool, errMsg string) {
	var outputArgs []byte
	if output != nil {
		outputArgs, _ = json.Marshal(output)
	}

	req := &workerservice.CompleteTaskRequest{
		TaskId:     taskId,
		OutputArgs: outputArgs,
		Success:    success,
		ErrorMsg:   []byte(errMsg),
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
