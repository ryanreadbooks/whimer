package service

import (
	"context"
	"errors"

	"github.com/ryanreadbooks/whimer/conductor/internal/biz"
	"github.com/ryanreadbooks/whimer/conductor/internal/biz/model"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

type WorkerService struct {
	bizz      *biz.Biz
	workerBiz *biz.WorkerBiz
	taskBiz   *biz.TaskBiz
}

func NewWorkerService(bizz *biz.Biz) *WorkerService {
	return &WorkerService{
		bizz:      bizz,
		workerBiz: bizz.WorkerBiz,
		taskBiz:   bizz.TaskBiz,
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
		// context.DeadlineExceeded 是正常的长轮询超时
		if errors.Is(err, context.DeadlineExceeded) {
			return &LongPollResponse{Task: nil}, nil
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
}

// CompleteTask Worker 完成任务上报
func (s *WorkerService) CompleteTask(ctx context.Context, req *CompleteTaskRequest) error {
	taskId, err := uuid.ParseString(req.TaskId)
	if err != nil {
		return xerror.ErrArgs.Msg("invalid task id")
	}

	err = s.bizz.Tx(ctx, func(ctx context.Context) error {
		return s.taskBiz.CompleteTask(ctx, taskId, req.Success, req.OutputArgs, req.ErrorMsg)
	})
	if err != nil {
		return xerror.Wrapf(err, "worker service complete task failed").WithCtx(ctx)
	}

	return nil
}
