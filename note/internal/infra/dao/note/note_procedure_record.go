package note

import (
	"context"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

const (
	procedureRecordTableName = "note_procedure_record"
)

var (
	procedureRecordFields    = xsql.GetFieldSlice(&ProcedureRecordPO{})
	procedureRecordInsFields = xsql.GetFieldSlice(&ProcedureRecordPO{})
)

type ProcedureRecordPO struct {
	Id            int64                 `db:"id"` // pk
	NoteId        int64                 `db:"note_id"`
	Protype       model.ProcedureType   `db:"protype"`
	TaskId        string                `db:"task_id"`
	Status        model.ProcedureStatus `db:"status"`
	Ctime         int64                 `db:"ctime"` // unix second
	Utime         int64                 `db:"utime"` // unix second
	CurRetry      int                   `db:"cur_retry"`
	MaxRetryCnt   int                   `db:"max_retry_cnt"`
	NextCheckTime int64                 `db:"next_check_time"` // unix second
}

func (ProcedureRecordPO) TableName() string {
	return procedureRecordTableName
}

func (p *ProcedureRecordPO) InsertValues() []any {
	return []any{
		p.NoteId,
		p.Protype,
		p.TaskId,
		p.Status,
		p.Ctime,
		p.Utime,
		p.CurRetry,
		p.MaxRetryCnt,
		p.NextCheckTime,
	}
}

type ProcedureRecordDao struct {
	db *xsql.DB
}

func NewProcedureRecordDao(db *xsql.DB) *ProcedureRecordDao {
	return &ProcedureRecordDao{db: db}
}

func (d *ProcedureRecordDao) Insert(
	ctx context.Context,
	record *ProcedureRecordPO,
) error {
	now := time.Now().Unix()
	if record.Ctime == 0 {
		record.Ctime = now
	}
	record.Utime = now

	sb := sqlbuilder.NewInsertBuilder()
	sb.InsertInto(procedureRecordTableName).
		Cols(procedureRecordInsFields...).
		Values(record.InsertValues()...)

	sql, args := sb.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xsql.ConvertError(err)
}

func (d *ProcedureRecordDao) UpdateTaskId(
	ctx context.Context,
	noteId int64,
	protype model.ProcedureType,
	taskId string,
) error {
	sb := sqlbuilder.NewUpdateBuilder()
	sb.Update(procedureRecordTableName).
		Set(
			sb.Assign("task_id", taskId),
			sb.Assign("utime", time.Now().Unix()),
		).
		Where(
			sb.EQ("note_id", noteId),
			sb.EQ("protype", protype),
		)

	sql, args := sb.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xsql.ConvertError(err)
}

func (d *ProcedureRecordDao) UpdateRecord(
	ctx context.Context,
	record *ProcedureRecordPO,
) error {
	sb := sqlbuilder.NewUpdateBuilder()
	sb.Update(procedureRecordTableName).
		Set(
			sb.Assign("task_id", record.TaskId),
			sb.Assign("status", record.Status),
			sb.Assign("cur_retry", record.CurRetry),
			sb.Assign("next_check_time", record.NextCheckTime),
			sb.Assign("utime", time.Now().Unix()),
		).
		Where(sb.EQ("id", record.Id))

	sql, args := sb.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xsql.ConvertError(err)
}

func (d *ProcedureRecordDao) UpdateStatus(
	ctx context.Context,
	noteId int64,
	protype model.ProcedureType,
	status model.ProcedureStatus,
) error {
	sb := sqlbuilder.NewUpdateBuilder()
	sb.Update(procedureRecordTableName).
		Set(
			sb.Assign("status", status),
			sb.Assign("utime", time.Now().Unix()),
		).
		Where(
			sb.EQ("note_id", noteId),
			sb.EQ("protype", protype),
		)

	sql, args := sb.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xsql.ConvertError(err)
}

func (d *ProcedureRecordDao) UpdateRetry(
	ctx context.Context,
	noteId int64,
	protype model.ProcedureType,
	nextCheckTime int64,
) error {
	sb := sqlbuilder.NewUpdateBuilder()
	sb.Update(procedureRecordTableName).
		Set(
			sb.Incr("cur_retry"),
			sb.Assign("next_check_time", nextCheckTime),
			sb.Assign("utime", time.Now().Unix()),
		).
		Where(
			sb.EQ("note_id", noteId),
			sb.EQ("protype", protype),
		)

	sql, args := sb.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xsql.ConvertError(err)
}

func (d *ProcedureRecordDao) Get(
	ctx context.Context,
	noteId int64,
	protype model.ProcedureType,
) (*ProcedureRecordPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(procedureRecordFields...).
		From(procedureRecordTableName).
		Where(
			sb.EQ("note_id", noteId),
			sb.EQ("protype", protype),
		)

	sql, args := sb.Build()
	record := new(ProcedureRecordPO)
	err := d.db.QueryRowCtx(ctx, record, sql, args...)
	return record, xsql.ConvertError(err)
}

func (d *ProcedureRecordDao) Delete(
	ctx context.Context,
	noteId int64,
	protype model.ProcedureType,
) error {
	sb := sqlbuilder.NewDeleteBuilder()
	sb.DeleteFrom(procedureRecordTableName).
		Where(
			sb.EQ("note_id", noteId),
			sb.EQ("protype", protype),
		)

	sql, args := sb.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xsql.ConvertError(err)
}

// 获取next_check_time最小的那条记录 且 status=特定status 且 protype=特定type
func (d *ProcedureRecordDao) GetEarliestNextCheckTimeRecord(
	ctx context.Context,
	status model.ProcedureStatus,
) (*ProcedureRecordPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(procedureRecordFields...).
		From(procedureRecordTableName).
		Where(
			sb.GTE("next_check_time", 0),
			sb.EQ("status", status),
		).
		OrderByAsc("next_check_time").
		OrderByAsc("id").
		Limit(1)

	sql, args := sb.Build()
	var record ProcedureRecordPO
	err := d.db.QueryRowCtx(ctx, &record, sql, args...)
	return &record, xsql.ConvertError(err)
}

func (d *ProcedureRecordDao) ListNextCheckTimeByRange(
	ctx context.Context,
	status model.ProcedureStatus,
	start int64,
	end int64,
	offsetId int64,
	limit int,
	shardIndex int,
	shardTotal int,
) ([]*ProcedureRecordPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Args.Add(shardTotal) // $1
	sb.Args.Add(shardIndex) // $2
	sb.Select(procedureRecordFields...).
		From(procedureRecordTableName).
		Where(
			sb.GTE("next_check_time", start),
			sb.LT("next_check_time", end),
			sb.EQ("status", status),
			sb.GT("id", offsetId),
			"MOD(id, $1) = $2", // 扫描单独分片内的数据
		).
		OrderByAsc("next_check_time").
		OrderByAsc("id").
		Limit(limit)

	sql, args := sb.Build()
	var records []*ProcedureRecordPO
	err := d.db.QueryRowsCtx(ctx, &records, sql, args...)

	return records, xsql.ConvertError(err)
}
