package dao

import (
	"context"

	"github.com/huandu/go-sqlbuilder"
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

func (d *TaskDao) GetById(ctx context.Context, id []byte) (*TaskPO, error) {
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

func (d *TaskDao) GetByNamespaceIdAndTaskType(ctx context.Context, namespaceId []byte, taskType string) ([]*TaskPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(taskPOFields...)
	sb.From(taskPOTableName)
	sb.Where(
		sb.Equal("namespace_id", namespaceId),
		sb.Equal("task_type", taskType),
	)

	sql, args := sb.Build()
	var pos []*TaskPO
	err := d.db.QueryRowsCtx(ctx, &pos, sql, args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return pos, nil
}

func (d *TaskDao) UpdateById(ctx context.Context, id []byte, po *TaskPO) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(taskPOTableName)
	ub.Set(
		ub.Assign("namespace_id", po.Namespace),
		ub.Assign("task_type", po.TaskType),
		ub.Assign("input_args", po.InputArgs),
		ub.Assign("ouput_args", po.OutputArgs),
		ub.Assign("callback_url", po.CallbackUrl),
		ub.Assign("state", po.State),
		ub.Assign("max_retry_cnt", po.MaxRetryCnt),
		ub.Assign("max_timeout_sec", po.MaxTimeoutSec),
		ub.Assign("utime", po.Utime),
		ub.Assign("version", po.Version+1), // 版本号自增
	)
	ub.Where(
		ub.Equal("id", id),
		ub.Equal("version", po.Version), // 乐观锁
	)

	sql, args := ub.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (d *TaskDao) DeleteById(ctx context.Context, id []byte) error {
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
