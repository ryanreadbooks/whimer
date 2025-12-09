package worker

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	taskv1 "github.com/ryanreadbooks/whimer/conductor/api/task/v1"
	workerv1 "github.com/ryanreadbooks/whimer/conductor/api/worker/v1"
	workerservice "github.com/ryanreadbooks/whimer/conductor/api/workerservice/v1"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/zeromicro/go-zero/zrpc"
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
	HostConf zrpc.RpcClientConf

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

	quitCh chan struct{}
	doneCh chan struct{}
	cancel context.CancelFunc // 用于取消正在进行的 LongPoll
}

// New 创建 Worker
func New(opts Options) (*Worker, error) {
	if opts.Concurrency <= 0 {
		opts.Concurrency = 1
	}
	// 初始化重试配置默认值
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

	cli, err := zrpc.NewClient(opts.HostConf)
	if err != nil {
		return nil, err
	}

	return &Worker{
		opts:     opts,
		client:   workerservice.NewWorkerServiceClient(cli.Conn()),
		handlers: make(map[string]Handler),
		quitCh:   make(chan struct{}),
		doneCh:   make(chan struct{}),
	}, nil
}

// MustNew 创建 Worker，失败则 panic
func MustNew(opts Options) *Worker {
	w, err := New(opts)
	if err != nil {
		panic(err)
	}
	return w
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

	// 创建可取消的 context，Stop 时取消以中断 LongPoll
	ctx, w.cancel = context.WithCancel(ctx)

	var wg sync.WaitGroup
	sem := make(chan struct{}, w.opts.Concurrency)

	for _, taskType := range taskTypes {
		taskType := taskType
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
	close(w.quitCh)
	if w.cancel != nil {
		w.cancel() // 取消 context，中断正在进行的 LongPoll
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

		// 获取信号量，控制并发
		select {
		case <-w.quitCh:
			return
		case <-ctx.Done():
			return
		case sem <- struct{}{}:
		}

		task, err := w.poll(ctx, taskType)
		if err != nil {
			// 长轮询正常超时不打印错误日志，直接继续轮询
			if !isTimeoutError(err) {
				xlog.Msg("poll task failed").Err(err).Errorx(ctx)
				// 出错时短暂等待后重试
				time.Sleep(time.Second)
			}
			<-sem
			continue
		}

		if task == nil {
			// 没有任务，释放信号量继续轮询
			<-sem
			continue
		}

		// 处理任务
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
	// 长轮询超时由 server 端控制
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
	// Accept task
	_, err := w.client.AcceptTask(ctx, &workerservice.AcceptTaskRequest{
		TaskId: task.Id,
	})
	if err != nil {
		xlog.Msg("accept task failed").Extras("taskId", task.Id).Err(err).Errorx(ctx)
		return
	}

	// Get handler
	w.mu.RLock()
	handler, ok := w.handlers[task.TaskType]
	w.mu.RUnlock()

	if !ok {
		xlog.Msg("no handler for task type").Extras("taskType", task.TaskType).Errorx(ctx)
		w.completeTask(ctx, task.Id, nil, false, "no handler for task type")
		return
	}

	// Execute handler
	result := w.safeExecute(ctx, task, handler)

	// Complete task
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

	// 指数退避重试
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

		// 等待退避时间
		time.Sleep(backoff)

		// 计算下次退避时间
		backoff = min(time.Duration(float64(backoff)*w.opts.ReportRetry.Multiplier), w.opts.ReportRetry.MaxBackoff)
	}
}

type panicError struct {
	v any
}

func (e *panicError) Error() string {
	return "handler panic"
}

// isTimeoutError 判断是否为正常的超时错误
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	// context 超时
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	// gRPC 超时
	if s, ok := status.FromError(err); ok {
		return s.Code() == codes.DeadlineExceeded
	}
	return false
}
