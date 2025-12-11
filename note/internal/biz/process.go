package biz

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/note/internal/infra"
	notedao "github.com/ryanreadbooks/whimer/note/internal/infra/dao/note"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

type NoteProcessBiz struct{}

func NewNoteProcessBiz() NoteProcessBiz {
	return NoteProcessBiz{}
}

func (b *NoteProcessBiz) CreateRecord(ctx context.Context, noteId int64, taskId string) (int64, error) {
	record := &notedao.ProcessRecordPO{
		NoteId: noteId,
		TaskId: taskId,
		Status: model.ProcessStatusProcessing,
	}

	id, err := infra.Dao().ProcessRecordDao.Insert(ctx, record)
	if err != nil {
		return 0, xerror.Wrapf(err, "biz create process record failed").
			WithExtra("noteId", noteId).
			WithExtra("taskId", taskId).
			WithCtx(ctx)
	}

	return id, nil
}

func (b *NoteProcessBiz) UpdateRecordStatus(ctx context.Context, taskId string, status model.ProcessStatus) error {
	err := infra.Dao().ProcessRecordDao.UpdateStatus(ctx, taskId, status)
	if err != nil {
		return xerror.Wrapf(err, "biz update process record status failed").
			WithExtra("taskId", taskId).
			WithExtra("status", status).
			WithCtx(ctx)
	}

	return nil
}

func (b *NoteProcessBiz) MarkRecordSuccess(ctx context.Context, taskId string) error {
	return b.UpdateRecordStatus(ctx, taskId, model.ProcessStatusSuccess)
}

func (b *NoteProcessBiz) MarkRecordFailed(ctx context.Context, taskId string) error {
	return b.UpdateRecordStatus(ctx, taskId, model.ProcessStatusFailed)
}

func (b *NoteProcessBiz) GetRecordByTaskId(ctx context.Context, taskId string) (*notedao.ProcessRecordPO, error) {
	record, err := infra.Dao().ProcessRecordDao.GetByTaskId(ctx, taskId)
	if err != nil {
		return nil, xerror.Wrapf(err, "biz get process record by task id failed").
			WithExtra("taskId", taskId).
			WithCtx(ctx)
	}

	return record, nil
}

func (b *NoteProcessBiz) GetRecordByNoteId(ctx context.Context, noteId int64) (*notedao.ProcessRecordPO, error) {
	record, err := infra.Dao().ProcessRecordDao.GetByNoteId(ctx, noteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "biz get process record by note id failed").
			WithExtra("noteId", noteId).
			WithCtx(ctx)
	}

	return record, nil
}

func (b *NoteProcessBiz) ListRecordsByNoteId(ctx context.Context, noteId int64) ([]*notedao.ProcessRecordPO, error) {
	records, err := infra.Dao().ProcessRecordDao.ListByNoteId(ctx, noteId)
	if err != nil {
		return nil, xerror.Wrapf(err, "biz list process records by note id failed").
			WithExtra("noteId", noteId).
			WithCtx(ctx)
	}

	return records, nil
}

func (b *NoteProcessBiz) DeleteRecordsByNoteId(ctx context.Context, noteId int64) error {
	err := infra.Dao().ProcessRecordDao.DeleteByNoteId(ctx, noteId)
	if err != nil {
		return xerror.Wrapf(err, "biz delete process records by note id failed").
			WithExtra("noteId", noteId).
			WithCtx(ctx)
	}

	return nil
}
