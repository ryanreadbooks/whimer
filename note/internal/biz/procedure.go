package biz

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xsql"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/data"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

type NoteProcedureBiz struct {
	data *data.Data
}

func NewNoteProcedureBiz(dt *data.Data) *NoteProcedureBiz {
	return &NoteProcedureBiz{
		data: dt,
	}
}

type CreateProcedureRecordReq struct {
	NoteId      int64
	Protype     model.ProcedureType
	TaskId      string
	MaxRetryCnt int
}

// 创建一条任务记录 如果存在且状态非处理中则更新
//
// 否则返回错误
func (b *NoteProcedureBiz) CreateRecord(
	ctx context.Context,
	req *CreateProcedureRecordReq,
) error {
	// 设定第一次检查时间
	curRecord, err := b.data.ProcedureRecord.GetForUpdate(ctx, req.NoteId, req.Protype)
	if err != nil {
		if !xsql.IsNoRecord(err) {
			return xerror.Wrapf(err, "biz get process record for update failed").
				WithExtras("noteId", req.NoteId, "protype", req.Protype).
				WithCtx(ctx)
		}
	}

	// record可能不存在 不存在就不需要检查状态
	if curRecord != nil && curRecord.Status == model.ProcessStatusProcessing {
		return xerror.Wrap(global.ErrNoteProcessing)
	}

	nextCheckTime := time.Now().Add(config.Conf.RetryConfig.ProcedureRetry.TaskRegister.RetryInterval).Unix()
	newRecord := &notedao.ProcedureRecordPO{
		NoteId:        req.NoteId,
		Protype:       req.Protype,
		TaskId:        req.TaskId,
		Status:        model.ProcessStatusProcessing,
		MaxRetryCnt:   req.MaxRetryCnt,
		NextCheckTime: nextCheckTime,
	}

	// 任务已经处理完就直接覆盖
	err = b.data.ProcedureRecord.Upsert(ctx, newRecord)
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
	err := b.data.ProcedureRecord.UpdateTaskId(ctx, noteId, protype, taskId)
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
	err := b.data.ProcedureRecord.UpdateStatus(ctx, noteId, protype, status)
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
	markFailure bool,
) error {
	var err error
	if markFailure {
		err = b.data.ProcedureRecord.UpdateRetryMarkFailure(ctx, noteId, protype, nextCheckTime)
	} else {
		err = b.data.ProcedureRecord.UpdateRetry(ctx, noteId, protype, nextCheckTime)
	}
	if err != nil {
		return xerror.Wrapf(err, "biz update retry failed").
			WithExtras("noteId", noteId, "protype", protype).
			WithCtx(ctx)
	}
	return nil
}

// 更新taskId cur_retry 和 next_check_time
func (b *NoteProcedureBiz) UpdateTaskIdRetryNextCheckTime(
	ctx context.Context,
	record *ProcedureRecord,
) error {
	err := b.data.ProcedureRecord.UpdateTaskIdRetryNextCheckTime(ctx, &notedao.ProcedureRecordPO{
		Id:            record.Id,
		NoteId:        record.NoteId,
		Protype:       record.Protype,
		TaskId:        record.TaskId,
		NextCheckTime: record.NextCheckTime,
		CurRetry:      record.CurRetry,
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
	record, err := b.data.ProcedureRecord.Get(ctx, noteId, protype)
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
	err := b.data.ProcedureRecord.Delete(ctx, noteId, protype)
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
	record, err := b.data.ProcedureRecord.GetEarliestNextCheckTimeRecord(ctx, status)
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
	records, err := b.data.ProcedureRecord.ListNextCheckTimeByRange(
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
