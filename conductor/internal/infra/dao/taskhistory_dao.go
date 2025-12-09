package dao

import (
	"context"

	"github.com/huandu/go-sqlbuilder"
	"github.com/ryanreadbooks/whimer/misc/uuid"
	"github.com/ryanreadbooks/whimer/misc/xsql"
)

type TaskHistoryDao struct {
	db *xsql.DB
}

func NewTaskHistoryDao(db *xsql.DB) *TaskHistoryDao {
	return &TaskHistoryDao{
		db: db,
	}
}

func (d *TaskHistoryDao) Insert(ctx context.Context, po *TaskHistoryPO) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(taskHistoryPOTableName)
	ib.Cols(taskHistoryPOFields...)
	ib.Values(po.Values()...)

	sql, args := ib.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (d *TaskHistoryDao) GetById(
	ctx context.Context,
	id int64) (*TaskHistoryPO, error) {

	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(taskHistoryPOFields...)
	sb.From(taskHistoryPOTableName)
	sb.Where(sb.Equal("id", id))

	sql, args := sb.Build()
	var po TaskHistoryPO
	err := d.db.QueryRowCtx(ctx, &po, sql, args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return &po, nil
}

func (d *TaskHistoryDao) GetByTaskId(
	ctx context.Context,
	taskId uuid.UUID) ([]*TaskHistoryPO, error) {

	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(taskHistoryPOFields...)
	sb.From(taskHistoryPOTableName)
	sb.Where(sb.Equal("task_id", taskId))
	sb.OrderByAsc("id")

	sql, args := sb.Build()
	var pos []*TaskHistoryPO
	err := d.db.QueryRowsCtx(ctx, &pos, sql, args...)
	if err != nil {
		return nil, xsql.ConvertError(err)
	}

	return pos, nil
}

func (d *TaskHistoryDao) UpdateById(
	ctx context.Context,
	id int64, po *TaskHistoryPO) error {

	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update(taskHistoryPOTableName)
	ub.Set(
		ub.Assign("task_id", po.TaskId),
		ub.Assign("state", po.State),
		ub.Assign("retry_cnt", po.RetryCnt),
		ub.Assign("ctime", po.Ctime),
	)
	ub.Where(ub.Equal("id", id))

	sql, args := ub.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

func (d *TaskHistoryDao) DeleteById(ctx context.Context, id int64) error {
	db := sqlbuilder.NewDeleteBuilder()
	db.DeleteFrom(taskHistoryPOTableName)
	db.Where(db.Equal("id", id))

	sql, args := db.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return xsql.ConvertError(err)
	}

	return nil
}

// GetMaxRetryCnt 获取任务的最大重试次数（从历史记录中获取）
func (d *TaskHistoryDao) GetMaxRetryCnt(ctx context.Context, taskId uuid.UUID) (int, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("COALESCE(MAX(retry_cnt), 0)")
	sb.From(taskHistoryPOTableName)
	sb.Where(sb.Equal("task_id", taskId))

	sql, args := sb.Build()
	var maxRetryCnt int
	err := d.db.QueryRowCtx(ctx, &maxRetryCnt, sql, args...)
	if err != nil {
		return 0, xsql.ConvertError(err)
	}

	return maxRetryCnt, nil
}
