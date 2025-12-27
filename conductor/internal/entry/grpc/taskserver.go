package grpc

import (
	"context"

	taskv1 "github.com/ryanreadbooks/whimer/conductor/api/task/v1"
	taskservice "github.com/ryanreadbooks/whimer/conductor/api/taskservice/v1"
	"github.com/ryanreadbooks/whimer/conductor/internal/service"
)

type TaskServiceServer struct {
	taskservice.UnimplementedTaskServiceServer

	srv *service.Service
}

func NewTaskServiceServer(srv *service.Service) *TaskServiceServer {
	return &TaskServiceServer{
		srv: srv,
	}
}

// RegisterTask 注册任务
func (s *TaskServiceServer) RegisterTask(ctx context.Context,
	in *taskservice.RegisterTaskRequest) (*taskservice.RegisterTaskResponse, error) {
	// 将 protobuf Timestamp 转换为 unix ms
	var expireTime int64
	if in.ExpireTime != nil {
		expireTime = in.ExpireTime.AsTime().UnixMilli()
	}

	resp, err := s.srv.TaskService.RegisterTask(ctx, &service.RegisterTaskReq{
		TaskType:    in.TaskType,
		Namespace:   in.Namespace,
		InputArgs:   in.InputArgs,
		CallbackUrl: in.CallbackUrl,
		MaxRetryCnt: in.MaxRetryCnt,
		ExpireTime:  expireTime,
	})
	if err != nil {
		return nil, err
	}

	return &taskservice.RegisterTaskResponse{
		TaskId: resp.TaskId,
	}, nil
}

// GetTask 获取任务
func (s *TaskServiceServer) GetTask(ctx context.Context,
	in *taskservice.GetTaskRequest) (*taskservice.GetTaskResponse, error) {
	task, err := s.srv.TaskService.GetTask(ctx, in.TaskId)
	if err != nil {
		return nil, err
	}

	return &taskservice.GetTaskResponse{
		Task: &taskv1.Task{
			Id:          task.Id.String(),
			Namespace:   task.Namespace,
			TaskType:    task.TaskType,
			InputArgs:   task.InputArgs,
			OutputArgs:  task.OutputArgs,
			CallbackUrl: task.CallbackUrl,
			State:       string(task.State.ExternalState()),
			MaxRetryCnt: task.MaxRetryCnt,
			ExpireTime:  task.ExpireTime,
			Ctime:       task.Ctime,
			Utime:       task.Utime,
			TraceId:     task.TraceId,
		},
	}, nil
}

// AbortTask 终止任务
func (s *TaskServiceServer) AbortTask(ctx context.Context,
	in *taskservice.AbortTaskRequest) (*taskservice.AbortTaskResponse, error) {
	err := s.srv.TaskService.AbortTask(ctx, in.TaskId)
	if err != nil {
		return nil, err
	}

	return &taskservice.AbortTaskResponse{}, nil
}
