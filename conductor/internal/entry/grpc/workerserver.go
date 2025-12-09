package grpc

import (
	"context"

	taskv1 "github.com/ryanreadbooks/whimer/conductor/api/task/v1"
	workerservice "github.com/ryanreadbooks/whimer/conductor/api/workerservice/v1"
	"github.com/ryanreadbooks/whimer/conductor/internal/service"
)

type WorkerServiceServer struct {
	workerservice.UnimplementedWorkerServiceServer

	srv *service.Service
}

func NewWorkerServiceServer(srv *service.Service) *WorkerServiceServer {
	return &WorkerServiceServer{
		srv: srv,
	}
}

// LongPoll Worker 长轮询获取任务
func (s *WorkerServiceServer) LongPoll(
	ctx context.Context,
	in *workerservice.LongPollRequest,
) (*workerservice.LongPollResponse, error) {
	var (
		workerId string
		taskType string
	)

	if in.Worker != nil {
		workerId = in.GetWorker().GetId()
		taskType = in.GetWorker().GetAbility().GetTaskType()
	}

	resp, err := s.srv.WorkerService.LongPoll(ctx, &service.LongPollRequest{
		WorkerId: workerId,
		TaskType: taskType,
	})
	if err != nil {
		return nil, err
	}

	// 没有任务（超时）
	if resp.Task == nil {
		return &workerservice.LongPollResponse{}, nil
	}

	return &workerservice.LongPollResponse{
		// 不返回 callbackurl 和 outputargs
		Task: &taskv1.Task{
			Id:          resp.Task.Id.String(),
			Namespace:   resp.Task.Namespace,
			TaskType:    resp.Task.TaskType,
			InputArgs:   resp.Task.InputArgs,
			State:       string(resp.Task.State),
			MaxRetryCnt: resp.Task.MaxRetryCnt,
			ExpireTime:  resp.Task.ExpireTime,
			Ctime:       resp.Task.Ctime,
			Utime:       resp.Task.Utime,
			TraceId:     resp.Task.TraceId,
		},
	}, nil
}

// AcceptTask Worker 接受任务
func (s *WorkerServiceServer) AcceptTask(
	ctx context.Context,
	in *workerservice.AcceptTaskRequest,
) (*workerservice.AcceptTaskResponse, error) {
	err := s.srv.WorkerService.AcceptTask(ctx, &service.AcceptTaskRequest{
		TaskId: in.TaskId,
	})
	if err != nil {
		return nil, err
	}

	return &workerservice.AcceptTaskResponse{}, nil
}

// CompleteTask Worker 完成任务上报
func (s *WorkerServiceServer) CompleteTask(
	ctx context.Context,
	in *workerservice.CompleteTaskRequest,
) (*workerservice.CompleteTaskResponse, error) {
	err := s.srv.WorkerService.CompleteTask(ctx, &service.CompleteTaskRequest{
		TaskId:     in.TaskId,
		Success:    in.Success,
		OutputArgs: in.OutputArgs,
		ErrorMsg:   in.ErrorMsg,
	})
	if err != nil {
		return nil, err
	}

	return &workerservice.CompleteTaskResponse{}, nil
}
