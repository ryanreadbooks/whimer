package data

import (
	"context"

	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

// ProcedureRecordData 流程记录数据层
type ProcedureRecordData struct {
	repo *notedao.ProcedureRecordRepo
}

func NewProcedureRecordData(repo *notedao.ProcedureRecordRepo) *ProcedureRecordData {
	return &ProcedureRecordData{
		repo: repo,
	}
}

// Insert 插入流程记录
func (d *ProcedureRecordData) Insert(ctx context.Context, record *notedao.ProcedureRecordPO) error {
	return d.repo.Insert(ctx, record)
}

// UpdateTaskId 更新流程记录的任务ID
func (d *ProcedureRecordData) UpdateTaskId(ctx context.Context, noteId int64, protype model.ProcedureType, taskId string) error {
	return d.repo.UpdateTaskId(ctx, noteId, protype, taskId)
}

// UpdateRecord 更新流程记录
func (d *ProcedureRecordData) UpdateTaskIdRetryNextCheckTime(ctx context.Context, record *notedao.ProcedureRecordPO) error {
	return d.repo.UpdateTaskIdRetryNextCheckTime(ctx, record)
}

// UpdateStatus 更新流程记录状态
func (d *ProcedureRecordData) UpdateStatus(ctx context.Context, noteId int64, protype model.ProcedureType, status model.ProcedureStatus) error {
	return d.repo.UpdateStatus(ctx, noteId, protype, status)
}

// UpdateRetry 更新流程记录重试次数
func (d *ProcedureRecordData) UpdateRetry(ctx context.Context, noteId int64, protype model.ProcedureType, nextCheckTime int64) error {
	return d.repo.UpdateRetry(ctx, noteId, protype, nextCheckTime)
}

// Get 获取流程记录
func (d *ProcedureRecordData) Get(ctx context.Context, noteId int64, protype model.ProcedureType) (*notedao.ProcedureRecordPO, error) {
	return d.repo.Get(ctx, noteId, protype)
}

// Delete 删除流程记录
func (d *ProcedureRecordData) Delete(ctx context.Context, noteId int64, protype model.ProcedureType) error {
	return d.repo.Delete(ctx, noteId, protype)
}

// GetEarliestNextCheckTimeRecord 获取最早需要检查的流程记录
func (d *ProcedureRecordData) GetEarliestNextCheckTimeRecord(ctx context.Context, status model.ProcedureStatus) (*notedao.ProcedureRecordPO, error) {
	return d.repo.GetEarliestNextCheckTimeRecord(ctx, status)
}

// ListNextCheckTimeByRange 获取指定时间范围内需要检查的流程记录
func (d *ProcedureRecordData) ListNextCheckTimeByRange(
	ctx context.Context,
	status model.ProcedureStatus,
	start int64,
	end int64,
	offsetId int64,
	limit int,
	shardIndex int,
	shardTotal int,
) ([]*notedao.ProcedureRecordPO, error) {
	return d.repo.ListNextCheckTimeByRange(ctx, status, start, end, offsetId, limit, shardIndex, shardTotal)
}
