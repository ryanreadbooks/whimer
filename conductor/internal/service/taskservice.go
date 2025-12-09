package service

import (
	"context"

	"github.com/ryanreadbooks/whimer/conductor/internal/biz"
	"github.com/ryanreadbooks/whimer/conductor/internal/biz/model"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

type TaskService struct {
	bizz         *biz.Biz
	namespaceBiz *biz.NamespaceBiz
	taskBiz      *biz.TaskBiz
}

func NewTaskService(bizz *biz.Biz) *TaskService {
	return &TaskService{
		bizz:         bizz,
		namespaceBiz: bizz.NamespaceBiz,
		taskBiz:      bizz.TaskBiz,
	}
}

type RegisterTaskReq struct {
	TaskType    string
	Namespace   string
	InputArgs   []byte
	CallbackUrl string
	MaxRetryCnt int64 // -1 无限重试, 0 不重试
	ExpireTime  int64 // 过期时间 unix ms
}

type RegisterTaskResp struct {
	TaskId string
}

// RegisterTask 注册任务
func (s *TaskService) RegisterTask(ctx context.Context, req *RegisterTaskReq) (*RegisterTaskResp, error) {
	// 校验 namespace 是否存在
	_, err := s.namespaceBiz.Get(ctx, req.Namespace)
	if err != nil {
		return nil, xerror.Wrapf(err, "namespace not found").
			WithExtra("namespace", req.Namespace).WithCtx(ctx)
	}

	var resp *biz.RegisterTaskResponse
	err = s.bizz.Tx(ctx, func(ctx context.Context) error {
		var txErr error
		resp, txErr = s.taskBiz.RegisterTask(ctx, &biz.RegisterTaskRequest{
			TaskType:    req.TaskType,
			Namespace:   req.Namespace,
			InputArgs:   req.InputArgs,
			CallbackUrl: req.CallbackUrl,
			MaxRetryCnt: req.MaxRetryCnt,
			ExpireTime:  req.ExpireTime,
		})
		return txErr
	})
	if err != nil {
		return nil, err
	}

	return &RegisterTaskResp{
		TaskId: resp.Task.Id.String(),
	}, nil
}

// GetTask 获取任务
func (s *TaskService) GetTask(ctx context.Context, taskId string) (*model.Task, error) {
	id, err := uuid.ParseString(taskId)
	if err != nil {
		return nil, xerror.ErrArgs.Msg("invalid task id")
	}

	task, err := s.taskBiz.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}

	return task, nil
}

// AbortTask 终止任务
func (s *TaskService) AbortTask(ctx context.Context, taskId string) error {
	id, err := uuid.ParseString(taskId)
	if err != nil {
		return xerror.ErrArgs.Msg("invalid task id")
	}

	// 只有 inited 和 dispatched 状态的任务可以被终止
	task, err := s.taskBiz.GetTask(ctx, id)
	if err != nil {
		return err
	}

	if task.State != model.TaskStateInited && task.State != model.TaskStateDispatched {
		return xerror.ErrArgs.Msg("task cannot be aborted")
	}

	return s.bizz.Tx(ctx, func(ctx context.Context) error {
		return s.taskBiz.UpdateTaskState(ctx, id, model.TaskStateAborted)
	})
}
