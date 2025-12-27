package dao

import (
	"context"

	"github.com/huandu/go-sqlbuilder"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type TaskDao struct {
	db *xsql.DB
}

func NewTaskDao(db *xsql.DB) *TaskDao {
	return &TaskDao{
		db: db,
	}
}

func (d *TaskDao) Insert(ctx context.Context, po *TaskPO) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(taskPOTableName)
	ib.Cols(taskPOFields...)
	ib.Values(po.Values()...)

	sql, args := ib.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (d *TaskDao) GetById(ctx context.Context, id uuid.UUID) (*TaskPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(taskPOFields...)
	sb.From(taskPOTableName)
	sb.Where(sb.Equal("id", id))

	sql, args := sb.Build()
	var po TaskPO
	err := d.db.QueryRowCtx(ctx, &po, sql, args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &po, nil
}

func (d *TaskDao) GetByNamespaceAndTaskTypeAndShard(ctx context.Context,
	namespace string,
	taskType string,
	shard int) ([]*TaskPO, error) {

	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(taskPOFields...)
	sb.From(taskPOTableName)
	sb.Where(
		sb.Equal("namespace", namespace),
		sb.Equal("task_type", taskType),
		sb.Equal("task_type_shard", shard),
	)

	sql, args := sb.Build()
	var pos []*TaskPO
	err := d.db.QueryRowsCtx(ctx, &pos, sql, args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return pos, nil
}

// ListTaskByState 按状态分页查询任务
func (d *TaskDao) ListTaskByState(
	ctx context.Context,
	state string,
	shardStart, shardEnd int, // [shardStart, shardEnd)
	limit int32,
	offset uuid.UUID,
) ([]*TaskPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(taskPOFields...)
	sb.From(taskPOTableName)
	sb.Where(
		sb.Equal("state", state),
		sb.GreaterEqualThan("task_type_shard", shardStart),
		sb.LessThan("task_type_shard", shardEnd),
		sb.GreaterThan("id", offset),
	)
	sb.OrderByAsc("id")
	sb.Limit(int(limit))

	sql, args := sb.Build()
	var pos []*TaskPO
	err := d.db.QueryRowsCtx(ctx, &pos, sql, args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}
	return pos, nil
}

// ListExpiredTasks 查询已过期的任务（非终态且 expire_time < now）
func (d *TaskDao) ListExpiredTasks(
	ctx context.Context,
	shardStart, shardEnd int,
	now int64,
	limit int32,
) ([]*TaskPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(taskPOFields...)
	sb.From(taskPOTableName)
	sb.Where(
		sb.In("state", "inited", "dispatched", "running", "pending_retry"),
		sb.GreaterEqualThan("task_type_shard", shardStart),
		sb.LessThan("task_type_shard", shardEnd),
		sb.GreaterThan("expire_time", 0),
		sb.LessThan("expire_time", now),
	)
	sb.OrderByAsc("id")
	sb.Limit(int(limit))

	sql, args := sb.Build()
	var pos []*TaskPO
	err := d.db.QueryRowsCtx(ctx, &pos, sql, args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}
	return pos, nil
}

func (d *TaskDao) UpdateById(ctx context.Context, id uuid.UUID, po *TaskPO) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(taskPOTableName)
	ub.Set(
		ub.Assign("namespace", po.Namespace),
		ub.Assign("task_type", po.TaskType),
		ub.Assign("task_type_shard", po.TaskTypeShard),
		ub.Assign("input_args", po.InputArgs),
		ub.Assign("output_args", po.OutputArgs),
		ub.Assign("callback_url", po.CallbackUrl),
		ub.Assign("state", po.State),
		ub.Assign("trace_id", po.TraceId),
		ub.Assign("utime", po.Utime),
		ub.Assign("max_retry_cnt", po.MaxRetryCnt),
		ub.Assign("expire_time", po.ExpireTime),
		ub.Assign("settings", po.Settings),
		ub.Assign("version", po.Version+1),
	)
	ub.Where(
		ub.Equal("id", id),
		ub.Equal("version", po.Version),
	)

	sql, args := ub.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (d *TaskDao) DeleteById(ctx context.Context, id uuid.UUID) error {
	db := sqlbuilder.NewDeleteBuilder()
	db.DeleteFrom(taskPOTableName)
	db.Where(db.Equal("id", id))

	sql, args := db.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (d *TaskDao) UpdateState(ctx context.Context, id uuid.UUID, state string, utime int64) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(taskPOTableName)
	ub.Set(
		ub.Assign("state", state),
		ub.Assign("utime", utime),
		ub.Incr("version"),
	)
	ub.Where(ub.Equal("id", id))

	sql, args := ub.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

// UpdateRetry 更新任务为重试状态，同时增加重试计数
func (d *TaskDao) UpdateRetry(ctx context.Context, id uuid.UUID, state string, utime int64) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(taskPOTableName)
	ub.Set(
		ub.Assign("state", state),
		ub.Assign("utime", utime),
		ub.Incr("cur_retry_cnt"),
		ub.Incr("version"),
	)
	ub.Where(ub.Equal("id", id))

	sql, args := ub.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (d *TaskDao) UpdateComplete(
	ctx context.Context,
	id uuid.UUID,
	state string,
	outputArgs []byte,
	utime int64,
) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(taskPOTableName)
	ub.Set(
		ub.Assign("state", state),
		ub.Assign("output_args", outputArgs),
		ub.Assign("utime", utime),
		ub.Incr("version"),
	)
	ub.Where(ub.Equal("id", id))

	sql, args := ub.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

// ListFailureTasks 查询失败状态且可重试的任务
// 条件：max_retry_cnt = -1（无限重试）或 cur_retry_cnt < max_retry_cnt（未达上限）
// func (d *TaskDao) ListFailureTasks(
// 	ctx context.Context,
// 	shardStart, shardEnd int,
// 	limit int32,
// 	offset uuid.UUID,
// ) ([]*TaskPO, error) {
// 	sb := sqlbuilder.NewSelectBuilder()
// 	sb.Select(taskPOFields...)
// 	sb.From(taskPOTableName)
// 	sb.Where(
// 		sb.Equal("state", "failure"),
// 		sb.GreaterEqualThan("task_type_shard", shardStart),
// 		sb.LessThan("task_type_shard", shardEnd),
// 		sb.GreaterThan("id", offset),
// 		// 可重试条件：无限重试(-1) 或 当前重试次数未达上限
// 		"(max_retry_cnt = -1 OR cur_retry_cnt < max_retry_cnt)",
// 	)
// 	sb.OrderByAsc("id")
// 	sb.Limit(int(limit))

// 	sql, args := sb.Build()
// 	var pos []*TaskPO
// 	err := d.db.QueryRowsCtx(ctx, &pos, sql, args...)
// 	if err != nil {
// 		return nil, xsql.ConvertError(err)
// 	}
// 	return pos, nil
// }
