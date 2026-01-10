package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/conductor/internal/biz"
	"github.com/ryanreadbooks/whimer/conductor/internal/biz/model"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

type WorkerService struct {
	bizz        *biz.Biz
	workerBiz   *biz.WorkerBiz
	taskBiz     *biz.TaskBiz
	callbackBiz *biz.CallbackBiz
}

func NewWorkerService(bizz *biz.Biz) *WorkerService {
	return &WorkerService{
		bizz:        bizz,
		workerBiz:   bizz.WorkerBiz,
		taskBiz:     bizz.TaskBiz,
		callbackBiz: bizz.CallbackBiz,
	}
}

type LongPollRequest struct {
	WorkerId string
	TaskType string
}

type LongPollResponse struct {
	Task *model.Task
}

// LongPoll Worker 长轮询获取任务
func (s *WorkerService) LongPoll(ctx context.Context, req *LongPollRequest) (*LongPollResponse, error) {
	task, err := s.workerBiz.WaitForTask(ctx, req.WorkerId, req.TaskType)
	if err != nil {
		// 正常的长轮询超时/取消
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return &LongPollResponse{Task: &model.Task{Id: uuid.EmptyUUID()}}, nil
		}
		return nil, xerror.Wrapf(err, "worker service long poll failed").WithCtx(ctx)
	}

	return &LongPollResponse{Task: task}, nil
}

type AcceptTaskRequest struct {
	TaskId string
}

// AcceptTask Worker 接受任务，更新状态为 running
func (s *WorkerService) AcceptTask(ctx context.Context, req *AcceptTaskRequest) error {
	taskId, err := uuid.ParseString(req.TaskId)
	if err != nil {
		return xerror.ErrArgs.Msg("invalid task id")
	}

	err = s.bizz.Tx(ctx, func(ctx context.Context) error {
		return s.taskBiz.UpdateTaskState(ctx, taskId, model.TaskStateRunning)
	})
	if err != nil {
		return xerror.Wrapf(err, "worker service accept task failed").WithCtx(ctx)
	}

	return nil
}

type CompleteTaskRequest struct {
	TaskId     string
	Success    bool
	OutputArgs []byte
	ErrorMsg   []byte
	Retryable  bool // 失败时是否可重试（由 worker 指定）
}

type ReportTaskRequest struct {
	TaskId   string
	Progress int64
}

type ReportTaskResponse struct {
	Aborted        bool
	NextReportTime time.Time // 建议下次上报时间（仅当未 aborted 时有效）
}

const defaultReportInterval = 10 * time.Second

// ReportTask Worker 上报当前任务进度
func (s *WorkerService) ReportTask(ctx context.Context, req *ReportTaskRequest) (*ReportTaskResponse, error) {
	taskId, err := uuid.ParseString(req.TaskId)
	if err != nil {
		return nil, xerror.ErrArgs.Msg("invalid task id")
	}

	task, err := s.taskBiz.GetTask(ctx, taskId)
	if err != nil {
		return nil, xerror.Wrapf(err, "worker service report task failed").WithCtx(ctx)
	}

	aborted := task.State == model.TaskStateAborted
	resp := &ReportTaskResponse{Aborted: aborted}

	// 如果任务未被终止，返回建议的下次上报时间
	if !aborted {
		resp.NextReportTime = time.Now().Add(defaultReportInterval)
	}

	return resp, nil
}

// CompleteTask Worker 完成任务上报
func (s *WorkerService) CompleteTask(ctx context.Context, req *CompleteTaskRequest) error {
	taskId, err := uuid.ParseString(req.TaskId)
	if err != nil {
		return xerror.ErrArgs.Msg("invalid task id")
	}

	// 先获取任务信息，用于后续回调
	task, err := s.taskBiz.GetTask(ctx, taskId)
	if err != nil {
		return xerror.Wrapf(err, "worker service get task failed").WithCtx(ctx)
	}

	// 如果任务已是终态（如 aborted, expired），不再更新状态
	if task.State.IsTerminal() {
		return nil
	}

	// 失败、可重试、且还有重试次数时，放入重试队列，不触发回调
	if !req.Success && req.Retryable && task.CanRetry() {
		err = s.bizz.Tx(ctx, func(ctx context.Context) error {
			return s.taskBiz.RetryTask(ctx, taskId)
		})
		if err != nil {
			return xerror.Wrapf(err, "worker service retry task failed").WithCtx(ctx)
		}
		return nil
	}

	// 成功或最终失败（重试次数用尽），更新为终态
	err = s.bizz.Tx(ctx, func(ctx context.Context) error {
		return s.taskBiz.CompleteTask(ctx, taskId, req.Success, req.OutputArgs, req.ErrorMsg)
	})
	if err != nil {
		return xerror.Wrapf(err, "worker service complete task failed").WithCtx(ctx)
	}

	// 只有终态才触发回调
	if task.CallbackUrl != "" {
		state := model.TaskStateSuccess
		if !req.Success {
			state = model.TaskStateFailure
		}

		s.callbackBiz.TriggerCallback(ctx, task.CallbackUrl, &biz.CallbackPayload{
			TaskId:      taskId.String(),
			Namespace:   task.Namespace,
			TaskType:    task.TaskType,
			State:       state,
			OutputArgs:  json.RawMessage(req.OutputArgs),
			ErrorMsg:    string(req.ErrorMsg),
			TraceId:     task.TraceId,
			CompletedAt: time.Now().UnixMilli(),
		})
	}

	return nil
}
