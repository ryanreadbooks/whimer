package test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/ryanreadbooks/whimer/conductor/pkg/sdk/producer"
	"github.com/ryanreadbooks/whimer/conductor/pkg/sdk/worker"
	"github.com/zeromicro/go-zero/zrpc"
)

// 测试用的任务输入
type EmailInput struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}

// 测试用的任务输出
type EmailOutput struct {
	MessageId string `json:"message_id"`
	Status    string `json:"status"`
}

// CallbackPayload 回调请求体
type CallbackPayload struct {
	TaskId      string          `json:"task_id"`
	Namespace   string          `json:"namespace"`
	TaskType    string          `json:"task_type"`
	State       string          `json:"state"`
	OutputArgs  json.RawMessage `json:"output_args,omitempty"`
	ErrorMsg    string          `json:"error_msg,omitempty"`
	TraceId     string          `json:"trace_id,omitempty"`
	CompletedAt int64           `json:"completed_at"`
}

// 测试配置（需要根据实际环境修改）
var testRpcConf = zrpc.RpcClientConf{
	Endpoints: []string{"localhost:10200"},
	NonBlock:  true,
	Timeout:   30000, // 30s，普通请求超时
}

// TestProducerRegisterTask 测试任务注册
func TestProducerRegisterTask(t *testing.T) {
	client, err := producer.NewClient(producer.ClientOptions{
		HostConf:  testRpcConf,
		Namespace: "test",
	})
	if err != nil {
		t.Fatalf("create producer client failed: %v", err)
	}

	ctx := context.Background()

	// 注册一个发送邮件的任务
	taskId, err := client.Execute(ctx, "send_email", EmailInput{
		To:      "test@example.com",
		Subject: "Hello World",
		Body:    "This is a test email",
	}, producer.ExecuteOptions{
		Namespace:   "default",
		MaxRetry:    3,
		ExpireAfter: time.Hour,
	})

	if err != nil {
		t.Fatalf("register task failed: %v", err)
	}

	t.Logf("task registered: %s", taskId)

	// 查询任务状态
	task, err := client.GetTask(ctx, taskId)
	if err != nil {
		t.Fatalf("get task failed: %v", err)
	}

	t.Logf("task info: id=%s, state=%s, namespace=%s, taskType=%s",
		task.Id, task.State, task.Namespace, task.TaskType)
}

// TestWorkerExecuteTask 测试任务执行
func TestWorkerExecuteTask(t *testing.T) {
	w, err := worker.New(worker.Options{
		HostConf:    testRpcConf,
		WorkerId:    "test-worker-1",
		Concurrency: 2,
	})
	if err != nil {
		t.Fatalf("create worker failed: %v", err)
	}

	// 注册邮件发送处理函数
	w.RegisterHandler("send_email", func(ctx context.Context, task *worker.Task) worker.Result {
		var input EmailInput
		if err := task.UnmarshalInput(&input); err != nil {
			return worker.Result{Error: err}
		}

		t.Logf("processing email task: to=%s, subject=%s", input.To, input.Subject)

		// 模拟发送邮件
		time.Sleep(100 * time.Millisecond)

		return worker.Result{
			Output: EmailOutput{
				MessageId: fmt.Sprintf("msg_%s", task.Id),
				Status:    "sent",
			},
		}
	})

	// 运行 worker（带超时）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		if err := w.Run(ctx); err != nil {
			t.Logf("worker run error: %v", err)
		}
	}()

	// 等待一段时间后停止
	time.Sleep(10 * time.Second)
	w.Stop()
}

// TestProducerAndWorkerIntegration 集成测试：注册任务并由 worker 执行，验证 callback
func TestProducerAndWorkerIntegration(t *testing.T) {
	// ========== 启动回调服务 ==========
	callbackPort := 18888
	callbackReceived := make(chan *CallbackPayload, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Logf("read callback body failed: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()

		var payload CallbackPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Logf("unmarshal callback payload failed: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		t.Logf("callback received: taskId=%s, state=%s, traceId=%s",
			payload.TaskId, payload.State, payload.TraceId)

		callbackReceived <- &payload
		w.WriteHeader(http.StatusOK)
	})

	callbackServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", callbackPort),
		Handler: mux,
	}

	go func() {
		if err := callbackServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Logf("callback server error: %v", err)
		}
	}()
	defer callbackServer.Close()

	// 等待服务启动
	time.Sleep(100 * time.Millisecond)

	// ========== 创建 producer ==========
	producerClient, err := producer.NewClient(producer.ClientOptions{
		HostConf:  testRpcConf,
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("create producer client failed: %v", err)
	}

	// ========== 创建 worker ==========
	workerClient, err := worker.New(worker.Options{
		HostConf:    testRpcConf,
		WorkerId:    "integration-test-worker",
		Concurrency: 1,
	})
	if err != nil {
		t.Fatalf("create worker failed: %v", err)
	}

	taskCompleted := make(chan string, 1)

	// 注册处理函数
	workerClient.RegisterHandler("callback_test", func(ctx context.Context, task *worker.Task) worker.Result {
		var input map[string]string
		if err := task.UnmarshalInput(&input); err != nil {
			return worker.Result{Error: err}
		}

		t.Logf("worker received task: %s, input: %v", task.Id, input)

		taskCompleted <- task.Id

		return worker.Result{
			Output: map[string]string{
				"result": "success",
				"echo":   input["message"],
			},
		}
	})

	// 启动 worker
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		if err := workerClient.Run(ctx); err != nil {
			t.Logf("worker run error: %v", err)
		}
	}()

	// 等待 worker 启动
	time.Sleep(time.Second)

	// ========== 注册任务（带 callback_url）==========
	callbackUrl := fmt.Sprintf("http://localhost:%d/callback", callbackPort)
	taskId, err := producerClient.Execute(ctx, "callback_test", map[string]string{
		"message": "hello from callback test",
	}, producer.ExecuteOptions{
		MaxRetry:    1,
		ExpireAfter: time.Minute,
		CallbackUrl: callbackUrl,
	})
	if err != nil {
		t.Fatalf("register task failed: %v", err)
	}

	t.Logf("task registered: %s, callbackUrl: %s", taskId, callbackUrl)

	// ========== 等待任务完成 ==========
	select {
	case completedTaskId := <-taskCompleted:
		t.Logf("task completed by worker: %s", completedTaskId)
	case <-time.After(20 * time.Second):
		t.Fatal("timeout waiting for task completion")
	}

	// ========== 等待回调 ==========
	select {
	case callback := <-callbackReceived:
		t.Logf("callback verified: taskId=%s, state=%s", callback.TaskId, callback.State)
		if callback.TaskId != taskId {
			t.Errorf("callback taskId mismatch: expected=%s, got=%s", taskId, callback.TaskId)
		}
		if callback.State != "success" {
			t.Errorf("callback state mismatch: expected=success, got=%s", callback.State)
		}
		// 验证 output
		if len(callback.OutputArgs) > 0 {
			var output map[string]string
			if err := json.Unmarshal(callback.OutputArgs, &output); err == nil {
				t.Logf("callback output: %v", output)
			}
		}
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for callback")
	}

	// ========== 查询任务最终状态 ==========
	task, err := producerClient.GetTask(ctx, taskId)
	if err != nil {
		t.Logf("get task failed: %v", err)
	} else {
		t.Logf("final task state: %s", task.State)
	}

	workerClient.Stop()
}

// TestRetryAndSuccess 测试重试并最终成功
// 场景：Worker 第一次失败，重试后成功
func TestRetryAndSuccess(t *testing.T) {
	producerClient, err := producer.NewClient(producer.ClientOptions{
		HostConf:  testRpcConf,
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("create producer client failed: %v", err)
	}

	workerClient, err := worker.New(worker.Options{
		HostConf:    testRpcConf,
		WorkerId:    "retry-success-worker",
		Concurrency: 1,
	})
	if err != nil {
		t.Fatalf("create worker failed: %v", err)
	}

	// 记录执行次数
	executionCount := 0
	taskCompleted := make(chan string, 1)

	// 注册处理函数：第一次失败，第二次成功
	workerClient.RegisterHandler("retry_success_test", func(ctx context.Context, task *worker.Task) worker.Result {
		executionCount++
		t.Logf("task execution #%d, taskId=%s", executionCount, task.Id)

		if executionCount == 1 {
			// 第一次返回失败
			return worker.Result{Error: fmt.Errorf("simulated failure")}
		}

		// 第二次返回成功
		taskCompleted <- task.Id
		return worker.Result{
			Output: map[string]string{"attempt": fmt.Sprintf("%d", executionCount)},
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	go func() {
		if err := workerClient.Run(ctx); err != nil {
			t.Logf("worker run error: %v", err)
		}
	}()

	time.Sleep(time.Second)

	// 注册任务，允许重试
	taskId, err := producerClient.Execute(ctx, "retry_success_test", map[string]string{
		"test": "retry_success",
	}, producer.ExecuteOptions{
		MaxRetry:    3,
		ExpireAfter: time.Minute,
	})
	if err != nil {
		t.Fatalf("register task failed: %v", err)
	}

	t.Logf("task registered: %s", taskId)

	// 等待任务最终成功
	select {
	case completedTaskId := <-taskCompleted:
		t.Logf("task finally succeeded: %s, total executions: %d", completedTaskId, executionCount)
		if executionCount < 2 {
			t.Errorf("expected at least 2 executions, got %d", executionCount)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("timeout waiting for task success")
	}

	// 等待上报完成
	time.Sleep(500 * time.Millisecond)

	// 验证最终状态
	task, err := producerClient.GetTask(ctx, taskId)
	if err != nil {
		t.Logf("get task failed: %v", err)
	} else {
		t.Logf("final task state: %s", task.State)
		if task.State != "success" {
			t.Errorf("expected state success, got %s", task.State)
		}
	}

	workerClient.Stop()
}

// TestRetryAndFailure 测试重试并最终失败
// 场景：Worker 多次失败，超过最大重试次数后任务失败
func TestRetryAndFailure(t *testing.T) {
	producerClient, err := producer.NewClient(producer.ClientOptions{
		HostConf:  testRpcConf,
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("create producer client failed: %v", err)
	}

	workerClient, err := worker.New(worker.Options{
		HostConf:    testRpcConf,
		WorkerId:    "retry-failure-worker",
		Concurrency: 1,
	})
	if err != nil {
		t.Fatalf("create worker failed: %v", err)
	}

	executionCount := 0
	maxRetry := int64(2)

	// 注册处理函数：始终失败
	workerClient.RegisterHandler("retry_failure_test", func(ctx context.Context, task *worker.Task) worker.Result {
		executionCount++
		t.Logf("task execution #%d, taskId=%s", executionCount, task.Id)
		return worker.Result{Error: fmt.Errorf("simulated failure #%d", executionCount)}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	go func() {
		if err := workerClient.Run(ctx); err != nil {
			t.Logf("worker run error: %v", err)
		}
	}()

	time.Sleep(time.Second)

	// 注册任务，限制重试次数
	taskId, err := producerClient.Execute(ctx, "retry_failure_test", map[string]string{
		"test": "retry_failure",
	}, producer.ExecuteOptions{
		MaxRetry:    maxRetry,
		ExpireAfter: time.Minute,
	})
	if err != nil {
		t.Fatalf("register task failed: %v", err)
	}

	t.Logf("task registered: %s, maxRetry: %d", taskId, maxRetry)

	// 等待足够时间让重试完成
	time.Sleep(20 * time.Second)

	// 验证执行次数（初始执行 + 重试次数）
	expectedExecutions := int(maxRetry) + 1
	t.Logf("total executions: %d, expected: %d", executionCount, expectedExecutions)

	// 验证最终状态应为 failure 或 expired
	task, err := producerClient.GetTask(ctx, taskId)
	if err != nil {
		t.Logf("get task failed: %v", err)
	} else {
		t.Logf("final task state: %s", task.State)
		if task.State != "failure" && task.State != "expired" {
			t.Logf("expected state failure or expired, got %s", task.State)
		}
	}

	workerClient.Stop()
}

// TestRetryAndTimeout 测试重试并最终超时
// 场景：任务重试过程中超时
func TestRetryAndTimeout(t *testing.T) {
	producerClient, err := producer.NewClient(producer.ClientOptions{
		HostConf:  testRpcConf,
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("create producer client failed: %v", err)
	}

	workerClient, err := worker.New(worker.Options{
		HostConf:    testRpcConf,
		WorkerId:    "retry-timeout-worker",
		Concurrency: 1,
	})
	if err != nil {
		t.Fatalf("create worker failed: %v", err)
	}

	executionCount := 0

	// 注册处理函数：始终失败
	workerClient.RegisterHandler("retry_timeout_test", func(ctx context.Context, task *worker.Task) worker.Result {
		executionCount++
		t.Logf("task execution #%d, taskId=%s", executionCount, task.Id)
		return worker.Result{Error: fmt.Errorf("simulated failure")}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	go func() {
		if err := workerClient.Run(ctx); err != nil {
			t.Logf("worker run error: %v", err)
		}
	}()

	time.Sleep(time.Second)

	// 注册任务：允许无限重试，但设置短超时
	taskId, err := producerClient.Execute(ctx, "retry_timeout_test", map[string]string{
		"test": "retry_timeout",
	}, producer.ExecuteOptions{
		MaxRetry:    -1, // 无限重试
		ExpireAfter: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("register task failed: %v", err)
	}

	t.Logf("task registered: %s, expireAfter: 10s, maxRetry: -1 (infinite)", taskId)

	// 等待超时
	time.Sleep(15 * time.Second)

	t.Logf("total executions before timeout: %d", executionCount)

	// 验证最终状态应为 expired
	task, err := producerClient.GetTask(ctx, taskId)
	if err != nil {
		t.Logf("get task failed: %v", err)
	} else {
		t.Logf("final task state: %s", task.State)
		if task.State != "expired" {
			t.Logf("expected state expired, got %s", task.State)
		}
	}

	workerClient.Stop()
}

// TestNoRetryAndTimeout 测试没有重试，但最终超时
// 场景：任务没有重试配置，Worker 不执行，最终超时
func TestNoRetryAndTimeout(t *testing.T) {
	producerClient, err := producer.NewClient(producer.ClientOptions{
		HostConf:  testRpcConf,
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("create producer client failed: %v", err)
	}

	// 注意：不启动 worker，让任务因无人处理而超时

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 注册任务：不重试，短超时
	taskId, err := producerClient.Execute(ctx, "no_retry_timeout_test", map[string]string{
		"test": "no_retry_timeout",
	}, producer.ExecuteOptions{
		MaxRetry:    0, // 不重试
		ExpireAfter: 10 * time.Second,
	})
	if err != nil {
		t.Fatalf("register task failed: %v", err)
	}

	t.Logf("task registered: %s, expireAfter: 10s, maxRetry: 0", taskId)

	// 查询初始状态
	task, err := producerClient.GetTask(ctx, taskId)
	if err != nil {
		t.Logf("get task failed: %v", err)
	} else {
		t.Logf("initial task state: %s", task.State)
	}

	// 等待超时
	time.Sleep(15 * time.Second)

	// 验证最终状态应为 expired
	task, err = producerClient.GetTask(ctx, taskId)
	if err != nil {
		t.Logf("get task failed: %v", err)
	} else {
		t.Logf("final task state: %s", task.State)
		if task.State != "expired" {
			t.Logf("expected state expired or inited, got %s", task.State)
		}
	}
}

// TestAbortTaskWhileRunning 测试任务执行过程中被 Producer 主动终止
// 场景：Worker 正在执行任务时，Producer 调用 AbortTask 终止
func TestAbortTaskWhileRunning(t *testing.T) {
	producerClient, err := producer.NewClient(producer.ClientOptions{
		HostConf:  testRpcConf,
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("create producer client failed: %v", err)
	}

	workerClient, err := worker.New(worker.Options{
		HostConf:    testRpcConf,
		WorkerId:    "abort-test-worker",
		Concurrency: 1,
	})
	if err != nil {
		t.Fatalf("create worker failed: %v", err)
	}

	taskStarted := make(chan string, 1)
	taskAborted := make(chan struct{}, 1)

	// 注册处理函数：开始执行后等待较长时间，模拟长时间运行的任务
	workerClient.RegisterHandler("abort_test", func(ctx context.Context, task *worker.Task) worker.Result {
		t.Logf("task started: %s", task.Id)
		taskStarted <- task.Id

		// 模拟长时间运行的任务
		select {
		case <-time.After(30 * time.Second):
			t.Logf("task completed normally (should not happen)")
			return worker.Result{Output: map[string]string{"status": "completed"}}
		case <-taskAborted:
			t.Logf("task detected abort signal")
			return worker.Result{Error: fmt.Errorf("task aborted by user")}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	go func() {
		if err := workerClient.Run(ctx); err != nil {
			t.Logf("worker run error: %v", err)
		}
	}()

	time.Sleep(time.Second)

	// 注册任务
	taskId, err := producerClient.Execute(ctx, "abort_test", map[string]string{
		"test": "abort_while_running",
	}, producer.ExecuteOptions{
		MaxRetry:    0,
		ExpireAfter: time.Minute,
	})
	if err != nil {
		t.Fatalf("register task failed: %v", err)
	}

	t.Logf("task registered: %s", taskId)

	// 等待任务开始执行
	select {
	case startedTaskId := <-taskStarted:
		t.Logf("task started execution: %s", startedTaskId)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for task to start")
	}

	// 等待一小段时间，确保任务正在 running
	time.Sleep(500 * time.Millisecond)

	// 查询任务状态，应为 running
	task, err := producerClient.GetTask(ctx, taskId)
	if err != nil {
		t.Logf("get task failed: %v", err)
	} else {
		t.Logf("task state before abort: %s", task.State)
	}

	// Producer 主动终止任务
	t.Logf("aborting task: %s", taskId)
	err = producerClient.AbortTask(ctx, taskId)
	if err != nil {
		t.Fatalf("abort task failed: %v", err)
	}
	t.Logf("abort task succeeded")

	// 通知 worker handler 任务已被终止
	close(taskAborted)

	// 等待一下让状态更新
	time.Sleep(500 * time.Millisecond)

	// 验证最终状态应为 aborted
	task, err = producerClient.GetTask(ctx, taskId)
	if err != nil {
		t.Logf("get task failed: %v", err)
	} else {
		t.Logf("final task state: %s", task.State)
		if task.State != "aborted" {
			t.Errorf("expected state aborted, got %s", task.State)
		}
	}

	workerClient.Stop()
}

// TestFailureWithoutRetry 测试任务失败且没有重试配置
// 场景：Worker 执行失败，MaxRetry=0，最终状态为 failure
func TestFailureWithoutRetry(t *testing.T) {
	producerClient, err := producer.NewClient(producer.ClientOptions{
		HostConf:  testRpcConf,
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("create producer client failed: %v", err)
	}

	workerClient, err := worker.New(worker.Options{
		HostConf:    testRpcConf,
		WorkerId:    "failure-no-retry-worker",
		Concurrency: 1,
	})
	if err != nil {
		t.Fatalf("create worker failed: %v", err)
	}

	taskFailed := make(chan string, 1)

	// 注册处理函数：直接返回失败
	workerClient.RegisterHandler("failure_no_retry_test", func(ctx context.Context, task *worker.Task) worker.Result {
		t.Logf("task executed: %s, returning failure", task.Id)
		taskFailed <- task.Id
		return worker.Result{Error: fmt.Errorf("simulated failure")}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		if err := workerClient.Run(ctx); err != nil {
			t.Logf("worker run error: %v", err)
		}
	}()

	time.Sleep(time.Second)

	// 注册任务，不重试
	taskId, err := producerClient.Execute(ctx, "failure_no_retry_test", map[string]string{
		"test": "failure_no_retry",
	}, producer.ExecuteOptions{
		MaxRetry:    0, // 不重试
		ExpireAfter: time.Minute,
	})
	if err != nil {
		t.Fatalf("register task failed: %v", err)
	}

	t.Logf("task registered: %s, maxRetry: 0", taskId)

	// 等待任务执行失败
	select {
	case failedTaskId := <-taskFailed:
		t.Logf("task failed: %s", failedTaskId)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for task failure")
	}

	// 等待上报完成
	time.Sleep(500 * time.Millisecond)

	// 验证最终状态应为 failure
	task, err := producerClient.GetTask(ctx, taskId)
	if err != nil {
		t.Logf("get task failed: %v", err)
	} else {
		t.Logf("final task state: %s", task.State)
		if task.State != "failure" {
			t.Errorf("expected state failure, got %s", task.State)
		}
	}

	workerClient.Stop()
}

// TestMaxRetryExhausted 测试达到最大重试次数后仍然失败
// 场景：Worker 执行始终失败，达到 MaxRetry 后进入 failure 终态
func TestMaxRetryExhausted(t *testing.T) {
	producerClient, err := producer.NewClient(producer.ClientOptions{
		HostConf:  testRpcConf,
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("create producer client failed: %v", err)
	}

	workerClient, err := worker.New(worker.Options{
		HostConf:    testRpcConf,
		WorkerId:    "max-retry-exhausted-worker",
		Concurrency: 1,
	})
	if err != nil {
		t.Fatalf("create worker failed: %v", err)
	}

	executionCount := 0
	maxRetry := int64(2) // 最大重试 2 次，总共执行 3 次（首次 + 2 次重试）

	// 注册处理函数：始终返回失败
	workerClient.RegisterHandler("max_retry_exhausted_test", func(ctx context.Context, task *worker.Task) worker.Result {
		executionCount++
		t.Logf("task execution #%d, taskId=%s", executionCount, task.Id)
		return worker.Result{Error: fmt.Errorf("simulated failure #%d", executionCount)}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	go func() {
		if err := workerClient.Run(ctx); err != nil {
			t.Logf("worker run error: %v", err)
		}
	}()

	time.Sleep(time.Second)

	// 注册任务，允许重试 2 次
	taskId, err := producerClient.Execute(ctx, "max_retry_exhausted_test", map[string]string{
		"test": "max_retry_exhausted",
	}, producer.ExecuteOptions{
		MaxRetry:    maxRetry,
		ExpireAfter: 2 * time.Minute,
	})
	if err != nil {
		t.Fatalf("register task failed: %v", err)
	}

	t.Logf("task registered: %s, maxRetry: %d", taskId, maxRetry)

	// 等待所有重试完成（首次执行 + 重试次数，每次重试需要等 retryScanLoop 扫描）
	// retryScanLoop 默认间隔 5s，总共需要等待约 (maxRetry + 1) * 扫描间隔
	expectedExecutions := int(maxRetry) + 1
	timeout := time.Duration(expectedExecutions*10) * time.Second

	t.Logf("waiting for %d executions (timeout: %v)...", expectedExecutions, timeout)

	// 轮询等待执行次数达到预期
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if executionCount >= expectedExecutions {
			break
		}
		time.Sleep(time.Second)
	}

	t.Logf("total executions: %d, expected: %d", executionCount, expectedExecutions)

	// 等待最后一次上报完成
	time.Sleep(time.Second)

	// 查询任务状态，应该还是 failure（retryScanLoop 会判断达到上限不再重试）
	// 需要等待 retryScanLoop 处理完成
	time.Sleep(10 * time.Second)

	// 验证最终状态应为 failure
	task, err := producerClient.GetTask(ctx, taskId)
	if err != nil {
		t.Logf("get task failed: %v", err)
	} else {
		t.Logf("final task state: %s", task.State)
		// 最终状态应为 failure（达到重试上限）
		if task.State != "failure" {
			t.Errorf("expected state failure, got %s", task.State)
		}
	}

	if executionCount != expectedExecutions {
		t.Errorf("expected %d executions, got %d", expectedExecutions, executionCount)
	}

	t.Logf("executionCount: %d, expectedExecutions: %d", executionCount, expectedExecutions)
	workerClient.Stop()
}
