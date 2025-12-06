package grpc

import (
	"context"

	taskv1 "github.com/ryanreadbooks/whimer/conductor/api/task/v1"
	taskservice "github.com/ryanreadbooks/whimer/conductor/api/taskservice/v1"
	"github.com/ryanreadbooks/whimer/conductor/internal/service"
)

type TaskServiceServer struct {
	taskservice.UnimplementedTaskServiceServer
}

func NewTaskServiceServer(srv *service.Service) *TaskServiceServer {
	return &TaskServiceServer{}
}

// 注册任务
func (s *TaskServiceServer) RegisterTask(ctx context.Context,
	in *taskservice.RegisterTaskRequest) (*taskservice.RegisterTaskResponse, error) {
	// TODO: implement

	return &taskservice.RegisterTaskResponse{}, nil
}

// 获取任务
func (s *TaskServiceServer) GetTask(ctx context.Context,
	in *taskservice.GetTaskRequest) (*taskservice.GetTaskResponse, error) {
	// TODO: implement

	return &taskservice.GetTaskResponse{
		Task: &taskv1.Task{},
	}, nil
}

// 终止任务
func (s *TaskServiceServer) AbortTask(ctx context.Context,
	in *taskservice.AbortTaskRequest) (*taskservice.AbortTaskResponse, error) {
	// TODO: implement

	return &taskservice.AbortTaskResponse{}, nil
}
