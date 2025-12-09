package biz

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ryanreadbooks/whimer/conductor/internal/biz/model"
	"github.com/ryanreadbooks/whimer/conductor/internal/infra/dao"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type TaskBiz struct {
	taskDao        *dao.TaskDao
	taskHistoryDao *dao.TaskHistoryDao
}

func NewTaskBiz(
	taskDao *dao.TaskDao,
	taskHistoryDao *dao.TaskHistoryDao,
) *TaskBiz {
	return &TaskBiz{
		taskDao:        taskDao,
		taskHistoryDao: taskHistoryDao,
	}
}

type RegisterTaskRequest struct {
	TaskType    string `json:"task_type"`
	Namespace   string `json:"namespace"`
	InputArgs   []byte `json:"input_args"`
	CallbackUrl string `json:"callback_url"`
	MaxRetryCnt int64  `json:"max_retry_cnt"` // -1 无限重试, 0 不重试
	ExpireTime  int64  `json:"expire_time"`   // 过期时间 unix ms
}

type RegisterTaskResponse struct {
	Task *model.Task
}

// RegisterTask 创建任务
func (b *TaskBiz) RegisterTask(
	ctx context.Context,
	req *RegisterTaskRequest) (*RegisterTaskResponse, error) {
	var (
		carrier    = propagation.MapCarrier{}
		propagator = otel.GetTextMapPropagator()
	)
	propagator.Inject(ctx, carrier)

	traceId := carrier.Get("traceparent")
	now := time.Now().UnixMilli()

	settings := model.TaskSettings{}
	settingBytes, err := json.Marshal(&settings)
	if err != nil {
		return nil, xerror.Wrapf(err, "task biz register task failed").WithCtx(ctx)
	}

	shard := model.CalculateShardHash(req.TaskType)
	taskPo := &dao.TaskPO{
		Id:            uuid.NewUUID(),
		Namespace:     req.Namespace,
		TaskType:      req.TaskType,
		TaskTypeShard: shard,
		InputArgs:     req.InputArgs,
		State:         string(model.TaskStateInited),
		CallbackUrl:   req.CallbackUrl,
		TraceId:       traceId,
		MaxRetryCnt:   req.MaxRetryCnt,
		ExpireTime:    req.ExpireTime,
		Settings:      settingBytes,
		Ctime:         now,
		Utime:         now,
		Version:       1,
	}
	err = b.taskDao.Insert(ctx, taskPo)
	if err != nil {
		return nil, xerror.Wrapf(err, "task biz register task failed").WithCtx(ctx)
	}

	taskHistoryPo := &dao.TaskHistoryPO{
		TaskId:   taskPo.Id,
		State:    string(model.TaskStateInited),
		RetryCnt: 0,
		Ctime:    now,
	}

	err = b.taskHistoryDao.Insert(ctx, taskHistoryPo)
	if err != nil {
		return nil, xerror.Wrapf(err, "task biz register task failed").WithCtx(ctx)
	}

	return &RegisterTaskResponse{
		Task: model.TaskFromPO(taskPo),
	}, nil
}

// GetTask 获取任务
func (b *TaskBiz) GetTask(
	ctx context.Context,
	taskId uuid.UUID) (*model.Task, error) {

	taskPo, err := b.taskDao.GetById(ctx, taskId)
	if err != nil {
		return nil, xerror.Wrapf(err, "task biz get task failed").WithCtx(ctx)
	}
	return model.TaskFromPO(taskPo), nil
}

// GetTaskHistorys 获取任务历史
func (b *TaskBiz) GetTaskHistorys(
	ctx context.Context,
	taskId uuid.UUID) ([]*model.TaskHistory, error) {

	pos, err := b.taskHistoryDao.GetByTaskId(ctx, taskId)
	if err != nil {
		return nil, xerror.Wrapf(err, "task biz get task historys failed").WithCtx(ctx)
	}
	taskHistories := make([]*model.TaskHistory, 0, len(pos))
	for _, taskHistoryPo := range pos {
		taskHistories = append(taskHistories, model.TaskHistoryFromPO(taskHistoryPo))
	}
	return taskHistories, nil
}

// GetTaskRetryCnt 获取任务当前的重试次数
func (b *TaskBiz) GetTaskRetryCnt(ctx context.Context, taskId uuid.UUID) (int, error) {
	cnt, err := b.taskHistoryDao.GetMaxRetryCnt(ctx, taskId)
	if err != nil {
		return 0, xerror.Wrapf(err, "task biz get task retry cnt failed").WithCtx(ctx)
	}
	return cnt, nil
}

// GetInitedTasks 获取创建成功但未被分配的任务
func (b *TaskBiz) GetInitedTasks(
	ctx context.Context,
	shardStart, shardEnd int, // [shardStart, shardEnd)
	limit int32,
	offset uuid.UUID,
) ([]*model.Task, error) {
	pos, err := b.taskDao.ListTaskByState(ctx,
		string(model.TaskStateInited),
		shardStart, shardEnd,
		limit, offset)
	if err != nil {
		return nil, xerror.Wrapf(err, "task biz get inited tasks failed").WithCtx(ctx)
	}
	tasks := make([]*model.Task, 0, len(pos))
	for _, taskPo := range pos {
		tasks = append(tasks, model.TaskFromPO(taskPo))
	}
	return tasks, nil
}

// GetExpiredTasks 获取已过期的任务
func (b *TaskBiz) GetExpiredTasks(
	ctx context.Context,
	shardStart, shardEnd int,
	limit int32,
) ([]*model.Task, error) {
	now := time.Now().UnixMilli()
	pos, err := b.taskDao.ListExpiredTasks(ctx, shardStart, shardEnd, now, limit)
	if err != nil {
		return nil, xerror.Wrapf(err, "task biz get expired tasks failed").WithCtx(ctx)
	}
	tasks := make([]*model.Task, 0, len(pos))
	for _, taskPo := range pos {
		tasks = append(tasks, model.TaskFromPO(taskPo))
	}
	return tasks, nil
}

// UpdateTaskState 更新任务状态
func (b *TaskBiz) UpdateTaskState(ctx context.Context, taskId uuid.UUID, state model.TaskState) error {
	now := time.Now().UnixMilli()
	err := b.taskDao.UpdateState(ctx, taskId, string(state), now)
	if err != nil {
		return xerror.Wrapf(err, "task biz update task state failed").
			WithExtra("taskId", taskId.String()).
			WithCtx(ctx)
	}

	taskHistoryPo := &dao.TaskHistoryPO{
		TaskId:   taskId,
		State:    string(state),
		RetryCnt: 0,
		Ctime:    now,
	}
	err = b.taskHistoryDao.Insert(ctx, taskHistoryPo)
	if err != nil {
		return xerror.Wrapf(err, "task biz insert task history failed").WithCtx(ctx)
	}

	return nil
}

// CompleteTask Worker 完成任务
func (b *TaskBiz) CompleteTask(
	ctx context.Context,
	taskId uuid.UUID,
	success bool,
	outputArgs []byte,
	errorMsg []byte,
) error {
	now := time.Now().UnixMilli()
	state := model.TaskStateSuccess
	if !success {
		state = model.TaskStateFailure
	}

	err := b.taskDao.UpdateComplete(ctx, taskId, string(state), outputArgs, now)
	if err != nil {
		return xerror.Wrapf(err, "task biz complete task failed").
			WithExtra("taskId", taskId.String()).
			WithCtx(ctx)
	}

	// 记录任务状态变更历史
	taskHistoryPo := &dao.TaskHistoryPO{
		TaskId:   taskId,
		State:    string(state),
		RetryCnt: 0,
		Ctime:    now,
	}
	err = b.taskHistoryDao.Insert(ctx, taskHistoryPo)
	if err != nil {
		return xerror.Wrapf(err, "task biz insert task history failed").WithCtx(ctx)
	}

	return nil
}

// RetryTask 将任务标记为待重试状态
func (b *TaskBiz) RetryTask(ctx context.Context, taskId uuid.UUID, retryCnt int) error {
	now := time.Now().UnixMilli()
	err := b.taskDao.UpdateState(ctx, taskId, string(model.TaskStatePendingRetry), now)
	if err != nil {
		return xerror.Wrapf(err, "task biz retry task failed").
			WithExtra("taskId", taskId.String()).
			WithCtx(ctx)
	}

	taskHistoryPo := &dao.TaskHistoryPO{
		TaskId:   taskId,
		State:    string(model.TaskStatePendingRetry),
		RetryCnt: retryCnt,
		Ctime:    now,
	}
	err = b.taskHistoryDao.Insert(ctx, taskHistoryPo)
	if err != nil {
		return xerror.Wrapf(err, "task biz insert task history failed").WithCtx(ctx)
	}

	return nil
}

// GetFailureTasks 获取失败状态的任务
func (b *TaskBiz) GetFailureTasks(
	ctx context.Context,
	shardStart, shardEnd int,
	limit int32,
	offset uuid.UUID,
) ([]*model.Task, error) {
	pos, err := b.taskDao.ListFailureTasks(ctx, shardStart, shardEnd, limit, offset)
	if err != nil {
		return nil, xerror.Wrapf(err, "task biz get failure tasks failed").WithCtx(ctx)
	}
	tasks := make([]*model.Task, 0, len(pos))
	for _, taskPo := range pos {
		tasks = append(tasks, model.TaskFromPO(taskPo))
	}
	return tasks, nil
}

// GetPendingRetryTasks 获取待重试状态的任务
func (b *TaskBiz) GetPendingRetryTasks(
	ctx context.Context,
	shardStart, shardEnd int,
	limit int32,
	offset uuid.UUID,
) ([]*model.Task, error) {
	pos, err := b.taskDao.ListTaskByState(ctx,
		string(model.TaskStatePendingRetry),
		shardStart, shardEnd,
		limit, offset)
	if err != nil {
		return nil, xerror.Wrapf(err, "task biz get pending retry tasks failed").WithCtx(ctx)
	}
	tasks := make([]*model.Task, 0, len(pos))
	for _, taskPo := range pos {
		tasks = append(tasks, model.TaskFromPO(taskPo))
	}
	return tasks, nil
}

// ExpireTask 将任务标记为过期
func (b *TaskBiz) ExpireTask(ctx context.Context, taskId uuid.UUID) error {
	return b.UpdateTaskState(ctx, taskId, model.TaskStateExpired)
}
