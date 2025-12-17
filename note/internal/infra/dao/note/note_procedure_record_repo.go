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
	procedureRecordInsFields = xsql.GetFieldSlice(&ProcedureRecordPO{}, "id")
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

// ProcedureRecordRepo 流程记录数据库仓储 - 纯数据库操作
type ProcedureRecordRepo struct {
	db *xsql.DB
}

func NewProcedureRecordRepo(db *xsql.DB) *ProcedureRecordRepo {
	return &ProcedureRecordRepo{db: db}
}

func (d *ProcedureRecordRepo) Upsert(
	ctx context.Context,
	record *ProcedureRecordPO,
	updateOnDup bool,
) error {
	now := time.Now().Unix()
	if record.Ctime == 0 {
		record.Ctime = now
	}
	record.Utime = now

	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto(procedureRecordTableName).
		Cols(procedureRecordInsFields...).
		Values(record.InsertValues()...)
	if updateOnDup {
		ib.SQL("ON DUPLICATE KEY UPDATE status=VALUES(status)," +
			"ctime=VALUES(ctime), " +
			"utime=VALUES(utime), " +
			"cur_retry=VALUES(cur_retry), " +
			"max_retry_cnt=VALUES(max_retry_cnt), " +
			"next_check_time=VALUES(next_check_time)")
	}

	sql, args := ib.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xsql.ConvertError(err)
}

func (d *ProcedureRecordRepo) UpdateTaskId(
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

func (d *ProcedureRecordRepo) UpdateTaskIdRetryNextCheckTime(
	ctx context.Context,
	record *ProcedureRecordPO,
) error {
	sb := sqlbuilder.NewUpdateBuilder()
	sb.Update(procedureRecordTableName).
		Set(
			sb.Assign("task_id", record.TaskId),
			sb.Assign("cur_retry", record.CurRetry),
			sb.Assign("next_check_time", record.NextCheckTime),
			sb.Assign("utime", time.Now().Unix()),
		).
		Where(sb.EQ("id", record.Id))

	sql, args := sb.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xsql.ConvertError(err)
}

func (d *ProcedureRecordRepo) UpdateStatus(
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

func (d *ProcedureRecordRepo) UpdateRetry(
	ctx context.Context,
	noteId int64,
	protype model.ProcedureType,
	nextCheckTime int64,
	markFailure bool,
) error {
	sb := sqlbuilder.NewUpdateBuilder()
	sb.Update(procedureRecordTableName).
		Set(
			sb.Incr("cur_retry"),
			sb.Assign("next_check_time", nextCheckTime),
			sb.Assign("utime", time.Now().Unix()),
		)
	if markFailure {
		sb.Set(
			sb.Assign("status", model.ProcessStatusFailed),
		)
	}

	sb.Where(
		sb.EQ("note_id", noteId),
		sb.EQ("protype", protype),
	)

	sql, args := sb.Build()
	_, err := d.db.ExecCtx(ctx, sql, args...)
	return xsql.ConvertError(err)
}

func (d *ProcedureRecordRepo) Get(
	ctx context.Context,
	noteId int64,
	protype model.ProcedureType,
	forUpdate bool,
) (*ProcedureRecordPO, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(procedureRecordFields...).
		From(procedureRecordTableName).
		Where(
			sb.EQ("note_id", noteId),
			sb.EQ("protype", protype),
		)
	if forUpdate {
		sb.ForUpdate()
	}

	sql, args := sb.Build()
	record := new(ProcedureRecordPO)
	err := d.db.QueryRowCtx(ctx, record, sql, args...)
	return record, xsql.ConvertError(err)
}

func (d *ProcedureRecordRepo) Delete(
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

// GetEarliestNextCheckTimeRecord 获取next_check_time最小的那条记录 且 status=特定status 且 protype=特定type
func (d *ProcedureRecordRepo) GetEarliestNextCheckTimeRecord(
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

func (d *ProcedureRecordRepo) ListNextCheckTimeByRange(
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
