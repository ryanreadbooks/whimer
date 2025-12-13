package biz

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

type NoteProcedureBiz struct {
	procedureDao *notedao.ProcedureRecordDao
}

func NewNoteProcedureBiz() NoteProcedureBiz {
	return NoteProcedureBiz{
		procedureDao: infra.Dao().ProcedureRecordDao,
	}
}

type CreateProcedureRecordReq struct {
	NoteId      int64
	Protype     model.ProcedureType
	TaskId      string
	MaxRetryCnt int
}

func (b *NoteProcedureBiz) CreateRecord(
	ctx context.Context,
	req *CreateProcedureRecordReq,
) error {
	// 设定第一次检查时间
	nextCheckTime := time.Now().Add(config.Conf.RetryConfig.ProcedureRetry.TaskRegister.RetryInterval).Unix()
	record := &notedao.ProcedureRecordPO{
		NoteId:        req.NoteId,
		Protype:       req.Protype,
		TaskId:        req.TaskId,
		Status:        model.ProcessStatusProcessing,
		MaxRetryCnt:   req.MaxRetryCnt,
		NextCheckTime: nextCheckTime,
	}

	err := b.procedureDao.Insert(ctx, record)
	if err != nil {
		return xerror.Wrapf(err, "biz create process record failed").
			WithExtras("noteId", req.NoteId, "protype", req.Protype).
			WithCtx(ctx)
	}

	return nil
}

func (b *NoteProcedureBiz) UpdateTaskId(
	ctx context.Context,
	noteId int64,
	protype model.ProcedureType,
	taskId string,
) error {
	err := b.procedureDao.UpdateTaskId(ctx, noteId, protype, taskId)
	if err != nil {
		return xerror.Wrapf(err, "biz update process record task id failed").
			WithExtras("noteId", noteId, "protype", protype, "taskId", taskId).
			WithCtx(ctx)
	}
	return nil
}

func (b *NoteProcedureBiz) UpdateStatus(
	ctx context.Context,
	noteId int64,
	protype model.ProcedureType,
	status model.ProcedureStatus,
) error {
	err := b.procedureDao.UpdateStatus(ctx, noteId, protype, status)
	if err != nil {
		return xerror.Wrapf(err, "biz update process record status failed").
			WithExtras("noteId", noteId, "protype", protype, "status", status).
			WithCtx(ctx)
	}
	return nil
}

func (b *NoteProcedureBiz) UpdateRetry(
	ctx context.Context,
	noteId int64,
	protype model.ProcedureType,
	nextCheckTime int64,
) error {
	err := b.procedureDao.UpdateRetry(ctx, noteId, protype, nextCheckTime)
	if err != nil {
		return xerror.Wrapf(err, "biz update retry failed").
			WithExtras("noteId", noteId, "protype", protype).
			WithCtx(ctx)
	}
	return nil
}

func (b *NoteProcedureBiz) UpdateRecord(
	ctx context.Context,
	record *ProcedureRecord,
) error {
	err := b.procedureDao.UpdateRecord(ctx, &notedao.ProcedureRecordPO{
		Id:            record.Id,
		NoteId:        record.NoteId,
		Protype:       record.Protype,
		TaskId:        record.TaskId,
		Status:        record.Status,
		NextCheckTime: record.NextCheckTime,
	})
	if err != nil {
		return xerror.Wrapf(err, "biz update process record failed").
			WithExtras("record", record).
			WithCtx(ctx)
	}
	return nil
}

func (b *NoteProcedureBiz) MarkSuccess(
	ctx context.Context,
	noteId int64,
	protype model.ProcedureType,
) error {
	return b.UpdateStatus(ctx, noteId, protype, model.ProcessStatusSuccess)
}

func (b *NoteProcedureBiz) MarkFailed(
	ctx context.Context,
	noteId int64,
	protype model.ProcedureType,
) error {
	return b.UpdateStatus(ctx, noteId, protype, model.ProcessStatusFailed)
}

func (b *NoteProcedureBiz) GetRecord(
	ctx context.Context,
	noteId int64,
	protype model.ProcedureType) (*ProcedureRecord, error) {
	record, err := b.procedureDao.Get(ctx, noteId, protype)
	if err != nil {
		return nil, xerror.Wrapf(err, "biz get process record failed").
			WithExtras("noteId", noteId, "protype", protype).
			WithCtx(ctx)
	}
	return ProcedureRecordFromPO(record), nil
}

func (b *NoteProcedureBiz) Delete(
	ctx context.Context,
	noteId int64,
	protype model.ProcedureType,
) error {
	err := b.procedureDao.Delete(ctx, noteId, protype)
	if err != nil {
		return xerror.Wrapf(err, "biz delete process records failed").
			WithExtra("noteId", noteId).
			WithCtx(ctx)
	}
	return nil
}

type ListProcessingRetryReq struct {
	CheckTimeOffset int64
	Offset          int64
	Limit           int
}

// 获取最早的未执行完成的下次检查时间记录
func (b *NoteProcedureBiz) GetEarliestScannedRecord(
	ctx context.Context,
	status model.ProcedureStatus) (*ProcedureRecord, error) {
	record, err := b.procedureDao.GetEarliestNextCheckTimeRecord(ctx, status)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return nil, nil
		}
		return nil, xerror.Wrapf(err, "biz get earliest next check time record failed").
			WithExtras("status", status).
			WithCtx(ctx)
	}

	return ProcedureRecordFromPO(record), nil
}

type ListRangeScannedRecordsReq struct {
	Status               model.ProcedureStatus
	RangeStart, RangeEnd int64
	OffsetId             int64
	Count                int
	ShardIdx, TotalShard int
}

func (b *NoteProcedureBiz) ListRangeScannedRecords(
	ctx context.Context,
	req *ListRangeScannedRecordsReq,
) (
	[]*ProcedureRecord, error,
) {
	records, err := b.procedureDao.ListNextCheckTimeByRange(
		ctx,
		req.Status,
		req.RangeStart, req.RangeEnd,
		req.OffsetId,
		req.Count,
		req.ShardIdx, req.TotalShard,
	)
	if err != nil {
		if xsql.IsNoRecord(err) {
			return []*ProcedureRecord{}, nil
		}

		return nil, xerror.Wrapf(err, "biz list range scanned records failed").
			WithExtra("req", req).
			WithCtx(ctx)
	}

	prs := make([]*ProcedureRecord, 0, len(records))
	for _, record := range records {
		prs = append(prs, ProcedureRecordFromPO(record))
	}

	return prs, nil
}
