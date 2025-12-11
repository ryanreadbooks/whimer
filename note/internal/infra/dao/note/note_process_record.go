package note

import (
	"context"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

const (
	processRecordTableName = "note_process_record"
)

var (
	processRecordFields    = xsql.GetFieldSlice(&ProcessRecordPO{})
	processRecordInsFields = xsql.GetFieldSlice(&ProcessRecordPO{}, "id")
)

type ProcessRecordPO struct {
	Id     int64               `db:"id"`
	NoteId int64               `db:"note_id"`
	TaskId string              `db:"task_id"`
	Status model.ProcessStatus `db:"status"`
	Ctime  int64               `db:"ctime"`
	Utime  int64               `db:"utime"`
}

func (ProcessRecordPO) TableName() string {
	return processRecordTableName
}

func (p *ProcessRecordPO) InsertValues() []any {
	return []any{
		p.NoteId,
		p.TaskId,
		p.Status,
		p.Ctime,
		p.Utime,
	}
}

type ProcessRecordDao struct {
	db *xsql.DB
}

func NewProcessRecordDao(db *xsql.DB) *ProcessRecordDao {
	return &ProcessRecordDao{db: db}
}

// Insert 插入处理记录
func (d *ProcessRecordDao) Insert(ctx context.Context, record *ProcessRecordPO) (int64, error) {
	now := time.Now().Unix()
	if record.Ctime == 0 {
		record.Ctime = now
	}
	record.Utime = now

	sb := sqlbuilder.NewInsertBuilder()
	sb.InsertInto(processRecordTableName).
		Cols(processRecordInsFields...).
		Values(record.InsertValues()...)

	sql, args := sb.Build()
	res, err := d.db.ExecCtx(ctx, sql, args...)
	if err != nil {
		return 0, xerror.Wrap(xsql.ConvertError(err))
	}

	id, _ := res.LastInsertId()
	return id, nil
}

// UpdateStatus 更新处理状态
func (d *ProcessRecordDao) UpdateStatus(
	ctx context.Context, taskId string, status model.ProcessStatus) error {
	sb := sqlbuilder.NewUpdateBuilder()
	sb.Update(processRecordTableName).
		Set(
			sb.Assign("status", status),
			sb.Assign("utime", time.Now().Unix()),
		).
		Where(sb.EQ("task_id", taskId))

	sql, args := sb.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xerror.Wrap(xsql.ConvertError(err))
}

// GetByTaskId 通过 task_id 查询
func (d *ProcessRecordDao) GetByTaskId(ctx context.Context, taskId string) (*ProcessRecordPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(processRecordFields...).
		From(processRecordTableName).
		Where(sb.EQ("task_id", taskId)).
		Limit(1)

	sql, args := sb.Build()
	var record ProcessRecordPO
	err := d.db.QueryRowCtx(ctx, &record, sql, args...)
	return &record, xerror.Wrap(xsql.ConvertError(err))
}

// GetByNoteId 通过 note_id 查询最新的处理记录
func (d *ProcessRecordDao) GetByNoteId(ctx context.Context, noteId int64) (*ProcessRecordPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(processRecordFields...).
		From(processRecordTableName).
		Where(sb.EQ("note_id", noteId)).
		OrderByDesc("id").
		Limit(1)

	sql, args := sb.Build()
	var record ProcessRecordPO
	err := d.db.QueryRowCtx(ctx, &record, sql, args...)
	return &record, xerror.Wrap(xsql.ConvertError(err))
}

// ListByNoteId 通过 note_id 查询所有处理记录
func (d *ProcessRecordDao) ListByNoteId(ctx context.Context, noteId int64) ([]*ProcessRecordPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(processRecordFields...).
		From(processRecordTableName).
		Where(sb.EQ("note_id", noteId)).
		OrderByDesc("id")

	sql, args := sb.Build()
	var records []*ProcessRecordPO
	err := d.db.QueryRowsCtx(ctx, &records, sql, args...)
	return records, xerror.Wrap(xsql.ConvertError(err))
}

// DeleteByNoteId 删除笔记的所有处理记录
func (d *ProcessRecordDao) DeleteByNoteId(ctx context.Context, noteId int64) error {
	sb := sqlbuilder.NewDeleteBuilder()
	sb.DeleteFrom(processRecordTableName).
		Where(sb.EQ("note_id", noteId))

	sql, args := sb.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xerror.Wrap(xsql.ConvertError(err))
}
