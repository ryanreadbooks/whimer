package test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ryanreadbooks/whimer/conductor/pkg/sdk/producer"
	"github.com/ryanreadbooks/whimer/conductor/pkg/sdk/worker"
)

// TestAbortWithTaskHandler 测试使用 TaskHandler 时的中断检测
// 场景：任务执行中，Producer 调用 AbortTask，Worker 通过心跳检测到 aborted 信号
func TestAbortWithTaskHandler(t *testing.T) {
	producerClient, err := producer.New(producer.ClientOptions{
		HostConf:  testRpcConf,
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("create producer client failed: %v", err)
	}

	workerClient, err := worker.New(worker.Options{
		HostConf:          testRpcConf,
		WorkerId:          "abort-taskhandler-worker",
		Concurrency:       1,
		HeartbeatInterval: 2 * time.Second, // 2秒心跳，加快测试速度
	})
	if err != nil {
		t.Fatalf("create worker failed: %v", err)
	}

	taskStarted := make(chan string, 1)
	taskAborted := make(chan string, 1)

	// 使用 TaskHandler 注册，通过 AbortCh 检测中断
	workerClient.RegisterTaskHandler("abort_taskhandler_test", func(tc worker.TaskContext) worker.Result {
		t.Logf("task started: %s", tc.Task().Id)
		taskStarted <- tc.Task().Id

		// 模拟长时间运行的任务，通过 AbortCh 检测中断
		for {
			select {
			case <-tc.AbortCh():
				t.Logf("task detected abort via AbortCh: %s", tc.Task().Id)
				taskAborted <- tc.Task().Id
				return worker.Result{Error: worker.ErrTaskAborted}
			case <-tc.Context().Done():
				t.Logf("task context cancelled: %s", tc.Task().Id)
				taskAborted <- tc.Task().Id
				return worker.Result{Error: worker.ErrTaskAborted}
			case <-time.After(100 * time.Millisecond):
				// 继续执行
				if tc.IsAborted() {
					t.Logf("task detected abort via IsAborted: %s", tc.Task().Id)
					taskAborted <- tc.Task().Id
					return worker.Result{Error: worker.ErrTaskAborted}
				}
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		if err := workerClient.Run(ctx); err != nil {
			t.Logf("worker run error: %v", err)
		}
	}()

	time.Sleep(time.Second)

	// 注册任务
	taskId, err := producerClient.Schedule(ctx, "abort_taskhandler_test", map[string]string{
		"test": "abort_with_taskhandler",
	}, producer.ScheduleOptions{
		MaxRetry:    0,
		ExpireAfter: time.Minute,
	})
	if err != nil {
		t.Fatalf("register task failed: %v", err)
	}
	t.Logf("task registered: %s", taskId)

	// 等待任务开始
	select {
	case <-taskStarted:
		t.Logf("task started execution")
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for task to start")
	}

	// 等待一个心跳周期，确保任务正在运行
	time.Sleep(3 * time.Second)

	// 终止任务
	t.Logf("aborting task: %s", taskId)
	err = producerClient.AbortTask(ctx, taskId)
	if err != nil {
		t.Fatalf("abort task failed: %v", err)
	}
	t.Logf("abort task succeeded")

	// 等待 Worker 检测到中断（需要等一个心跳周期）
	select {
	case <-taskAborted:
		t.Logf("worker detected abort signal")
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for worker to detect abort")
	}

	// 验证最终状态
	// 注意：ExternalState 将 aborted 对外展示为 failure
	time.Sleep(500 * time.Millisecond)
	task, err := producerClient.GetTask(ctx, taskId)
	if err != nil {
		t.Logf("get task failed: %v", err)
	} else {
		t.Logf("final task state: %s", task.State)
		// aborted 对外展示为 failure
		if task.State != "failure" {
			t.Errorf("expected state failure (aborted externally shown as failure), got %s", task.State)
		}
	}

	// 异步停止，避免阻塞（ctx cancel 后 LongPoll 需要时间响应）
	go workerClient.Stop()
}

// TestAbortWithProgressProvider 测试使用 ProgressProvider 时的进度上报和中断检测
func TestAbortWithProgressProvider(t *testing.T) {
	producerClient, err := producer.New(producer.ClientOptions{
		HostConf:  testRpcConf,
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("create producer client failed: %v", err)
	}

	workerClient, err := worker.New(worker.Options{
		HostConf:          testRpcConf,
		WorkerId:          "abort-progress-worker",
		Concurrency:       1,
		HeartbeatInterval: 1 * time.Second, // 1秒心跳
	})
	if err != nil {
		t.Fatalf("create worker failed: %v", err)
	}

	taskStarted := make(chan string, 1)
	taskAborted := make(chan string, 1)
	var currentProgress atomic.Int64

	// 使用 TaskHandler 并设置 ProgressProvider
	workerClient.RegisterTaskHandler("abort_progress_test", func(tc worker.TaskContext) worker.Result {
		t.Logf("task started: %s", tc.Task().Id)
		taskStarted <- tc.Task().Id

		// 设置进度提供者，心跳时自动上报进度
		tc.SetProgressProvider(worker.ProgressFunc(func() int64 {
			return currentProgress.Load()
		}))

		// 模拟逐步执行的任务
		for i := 0; i <= 100; i++ {
			currentProgress.Store(int64(i))

			select {
			case <-tc.AbortCh():
				t.Logf("task aborted at progress %d", i)
				taskAborted <- tc.Task().Id
				return worker.Result{Error: worker.ErrTaskAborted}
			case <-time.After(200 * time.Millisecond):
				// 继续执行
			}
		}

		return worker.Result{Output: map[string]string{"status": "completed"}}
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
	taskId, err := producerClient.Schedule(ctx, "abort_progress_test", map[string]string{
		"test": "abort_with_progress",
	}, producer.ScheduleOptions{
		MaxRetry:    0,
		ExpireAfter: time.Minute,
	})
	if err != nil {
		t.Fatalf("register task failed: %v", err)
	}
	t.Logf("task registered: %s", taskId)

	// 等待任务开始
	select {
	case <-taskStarted:
		t.Logf("task started execution")
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for task to start")
	}

	// 等待进度达到一定值
	time.Sleep(3 * time.Second)
	progress := currentProgress.Load()
	t.Logf("current progress before abort: %d", progress)

	// 终止任务
	t.Logf("aborting task: %s", taskId)
	err = producerClient.AbortTask(ctx, taskId)
	if err != nil {
		t.Fatalf("abort task failed: %v", err)
	}

	// 等待 Worker 检测到中断
	select {
	case <-taskAborted:
		finalProgress := currentProgress.Load()
		t.Logf("worker detected abort at progress %d", finalProgress)
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for worker to detect abort")
	}

	// 验证最终状态
	// 注意：ExternalState 将 aborted 对外展示为 failure
	time.Sleep(500 * time.Millisecond)
	task, err := producerClient.GetTask(ctx, taskId)
	if err != nil {
		t.Logf("get task failed: %v", err)
	} else {
		t.Logf("final task state: %s", task.State)
		if task.State != "failure" {
			t.Errorf("expected state failure (aborted externally shown as failure), got %s", task.State)
		}
	}

	// 异步停止，避免阻塞
	go workerClient.Stop()
}

// TestAbortWithSimpleHandler 测试使用简单 Handler 时的中断检测
// 场景：即使使用简单 Handler，心跳也会检测中断并取消 context
func TestAbortWithSimpleHandler(t *testing.T) {
	producerClient, err := producer.New(producer.ClientOptions{
		HostConf:  testRpcConf,
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("create producer client failed: %v", err)
	}

	workerClient, err := worker.New(worker.Options{
		HostConf:          testRpcConf,
		WorkerId:          "abort-simple-handler-worker",
		Concurrency:       1,
		HeartbeatInterval: 2 * time.Second,
	})
	if err != nil {
		t.Fatalf("create worker failed: %v", err)
	}

	taskStarted := make(chan string, 1)
	taskAborted := make(chan string, 1)

	// 使用简单 Handler，通过 ctx.Done() 检测中断
	workerClient.RegisterHandler("abort_simple_handler_test", func(ctx context.Context, task *worker.Task) worker.Result {
		t.Logf("task started: %s", task.Id)
		taskStarted <- task.Id

		// 模拟长时间运行，通过 context 检测中断
		for {
			select {
			case <-ctx.Done():
				t.Logf("task context cancelled: %s", task.Id)
				taskAborted <- task.Id
				return worker.Result{Error: worker.ErrTaskAborted}
			case <-time.After(100 * time.Millisecond):
				// 继续执行
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		if err := workerClient.Run(ctx); err != nil {
			t.Logf("worker run error: %v", err)
		}
	}()

	time.Sleep(time.Second)

	// 注册任务
	taskId, err := producerClient.Schedule(ctx, "abort_simple_handler_test", map[string]string{
		"test": "abort_simple_handler",
	}, producer.ScheduleOptions{
		MaxRetry:    0,
		ExpireAfter: time.Minute,
	})
	if err != nil {
		t.Fatalf("register task failed: %v", err)
	}
	t.Logf("task registered: %s", taskId)

	// 等待任务开始
	select {
	case <-taskStarted:
		t.Logf("task started execution")
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for task to start")
	}

	// 等待一个心跳周期
	time.Sleep(3 * time.Second)

	// 终止任务
	t.Logf("aborting task: %s", taskId)
	err = producerClient.AbortTask(ctx, taskId)
	if err != nil {
		t.Fatalf("abort task failed: %v", err)
	}

	// 等待 Worker 检测到中断
	select {
	case <-taskAborted:
		t.Logf("worker detected abort via context cancellation")
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for worker to detect abort")
	}

	// 验证最终状态
	// 注意：ExternalState 将 aborted 对外展示为 failure
	time.Sleep(500 * time.Millisecond)
	task, err := producerClient.GetTask(ctx, taskId)
	if err != nil {
		t.Logf("get task failed: %v", err)
	} else {
		t.Logf("final task state: %s", task.State)
		if task.State != "failure" {
			t.Errorf("expected state failure (aborted externally shown as failure), got %s", task.State)
		}
	}

	// 异步停止，避免阻塞
	go workerClient.Stop()
}

// TestManualReportProgress 测试手动上报进度并检测中断
func TestManualReportProgress(t *testing.T) {
	producerClient, err := producer.New(producer.ClientOptions{
		HostConf:  testRpcConf,
		Namespace: "default",
	})
	if err != nil {
		t.Fatalf("create producer client failed: %v", err)
	}

	workerClient, err := worker.New(worker.Options{
		HostConf:          testRpcConf,
		WorkerId:          "manual-progress-worker",
		Concurrency:       1,
		HeartbeatInterval: 10 * time.Second, // 长心跳，使用手动上报
	})
	if err != nil {
		t.Fatalf("create worker failed: %v", err)
	}

	taskStarted := make(chan string, 1)
	taskAborted := make(chan string, 1)

	// 手动调用 ReportProgress 上报进度
	workerClient.RegisterTaskHandler("manual_progress_test", func(tc worker.TaskContext) worker.Result {
		t.Logf("task started: %s", tc.Task().Id)
		taskStarted <- tc.Task().Id

		for i := 0; i <= 100; i++ {
			// 手动上报进度，同时检测中断
			if tc.ReportProgress(int64(i)) {
				t.Logf("task aborted at progress %d via ReportProgress", i)
				taskAborted <- tc.Task().Id
				return worker.Result{Error: worker.ErrTaskAborted}
			}

			time.Sleep(100 * time.Millisecond)
		}

		return worker.Result{Output: map[string]string{"status": "completed"}}
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
	taskId, err := producerClient.Schedule(ctx, "manual_progress_test", map[string]string{
		"test": "manual_progress",
	}, producer.ScheduleOptions{
		MaxRetry:    0,
		ExpireAfter: time.Minute,
	})
	if err != nil {
		t.Fatalf("register task failed: %v", err)
	}
	t.Logf("task registered: %s", taskId)

	// 等待任务开始
	select {
	case <-taskStarted:
		t.Logf("task started execution")
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for task to start")
	}

	// 等待一段时间
	time.Sleep(2 * time.Second)

	// 终止任务
	t.Logf("aborting task: %s", taskId)
	err = producerClient.AbortTask(ctx, taskId)
	if err != nil {
		t.Fatalf("abort task failed: %v", err)
	}

	// 等待 Worker 通过 ReportProgress 检测到中断
	select {
	case <-taskAborted:
		t.Logf("worker detected abort via ReportProgress")
	case <-time.After(10 * time.Second):
		t.Fatal("timeout waiting for worker to detect abort")
	}

	// 验证最终状态
	// 注意：ExternalState 将 aborted 对外展示为 failure
	time.Sleep(500 * time.Millisecond)
	task, err := producerClient.GetTask(ctx, taskId)
	if err != nil {
		t.Logf("get task failed: %v", err)
	} else {
		t.Logf("final task state: %s", task.State)
		if task.State != "failure" {
			t.Errorf("expected state failure (aborted externally shown as failure), got %s", task.State)
		}
	}

	// 异步停止，避免阻塞
	go workerClient.Stop()
}
