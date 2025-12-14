package procedure

import (
	"context"

	conductor "github.com/ryanreadbooks/whimer/conductor/pkg/sdk/producer"
	sdktask "github.com/ryanreadbooks/whimer/conductor/pkg/sdk/task"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/note/internal/model"
	"github.com/ryanreadbooks/whimer/note/internal/srv/assetprocess"
)

var _ Procedure = (*AssetProcedure)(nil)

// AssetProcedure 资源处理流程 负责笔记资源（图片、视频等）的处理
type AssetProcedure struct {
	bizz             *biz.Biz
	noteBiz          *biz.NoteBiz
	noteCreatorBiz   *biz.NoteCreatorBiz
	noteProcedureBiz *biz.NoteProcedureBiz

	conductorProducer *conductor.Client

	// 重试通用逻辑
	retryHelper *retryHelper
}

func NewAssetProcedure(bizz *biz.Biz) *AssetProcedure {
	return &AssetProcedure{
		bizz:              bizz,
		noteBiz:           bizz.Note,
		noteCreatorBiz:    bizz.Creator,
		noteProcedureBiz:  bizz.Procedure,
		conductorProducer: dep.GetConductProducer(),
		retryHelper:       newRetryHelper(bizz.Note, bizz.Procedure),
	}
}

func (p *AssetProcedure) Type() model.ProcedureType {
	return model.ProcedureTypeAssetProcess
}

func (p *AssetProcedure) PreStart(ctx context.Context, note *model.Note) error {
	err := p.noteCreatorBiz.SetNoteStateProcessing(ctx, note.NoteId)
	if err != nil {
		return xerror.Wrapf(err, "asset procedure set note state processing failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}
	return nil
}

func (p *AssetProcedure) Execute(ctx context.Context, note *model.Note) (string, error) {
	processor := assetprocess.NewProcessor(note.Type, p.bizz)
	taskId, err := processor.Process(ctx, note)
	if err != nil {
		return "", xerror.Wrapf(err, "asset procedure execute failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}
	return taskId, nil
}

func (p *AssetProcedure) OnSuccess(ctx context.Context, noteId int64, taskId string) error {
	record, err := p.noteProcedureBiz.GetRecord(ctx, noteId, model.ProcedureTypeAssetProcess)
	if err != nil {
		return xerror.Wrapf(err, "asset procedure get record failed").
			WithExtra("taskId", taskId).
			WithCtx(ctx)
	}

	err = p.bizz.Tx(ctx, func(ctx context.Context) error {
		if err := p.noteCreatorBiz.SetNoteStateProcessed(ctx, record.NoteId); err != nil {
			return xerror.Wrapf(err, "asset procedure set note state processed failed").
				WithExtra("noteId", record.NoteId).
				WithCtx(ctx)
		}

		if err := p.noteProcedureBiz.MarkSuccess(ctx, record.NoteId, record.Protype); err != nil {
			return xerror.Wrapf(err, "asset procedure mark record success failed").
				WithExtra("taskId", taskId).
				WithCtx(ctx)
		}

		return nil
	})
	if err != nil {
		xlog.Msg("asset procedure on success tx failed").
			Err(err).
			Extras("taskId", taskId).
			Errorx(ctx)
		return err
	}

	xlog.Msg("asset procedure on success completed").
		Extras("taskId", taskId, "noteId", record.NoteId).
		Infox(ctx)

	return nil
}

func (p *AssetProcedure) OnFailure(ctx context.Context, noteId int64, taskId string) error {
	record, err := p.noteProcedureBiz.GetRecord(ctx, noteId, model.ProcedureTypeAssetProcess)
	if err != nil {
		return xerror.Wrapf(err, "asset procedure get record failed").
			WithExtra("taskId", taskId).
			WithCtx(ctx)
	}

	err = p.bizz.Tx(ctx, func(ctx context.Context) error {
		if err := p.noteCreatorBiz.SetNoteStateProcessFailed(ctx, record.NoteId); err != nil {
			return xerror.Wrapf(err, "asset procedure set note state failed failed").
				WithExtra("noteId", record.NoteId).
				WithCtx(ctx)
		}

		if err := p.noteProcedureBiz.MarkFailed(ctx, record.NoteId, record.Protype); err != nil {
			return xerror.Wrapf(err, "asset procedure mark record failed failed").
				WithExtra("taskId", taskId).
				WithCtx(ctx)
		}

		return nil
	})
	if err != nil {
		xlog.Msg("asset procedure on failure tx failed").
			Err(err).
			Extras("taskId", taskId).
			Errorx(ctx)
		return err
	}

	xlog.Msg("asset procedure on failure completed").
		Extras("taskId", taskId, "noteId", record.NoteId).
		Infox(ctx)

	return nil
}

func (p *AssetProcedure) PollResult(ctx context.Context, taskId string) (bool, error) {
	task, err := p.conductorProducer.GetTask(ctx, taskId)
	if err != nil {
		return false, xerror.Wrapf(err, "asset procedure poll result failed").
			WithExtra("task_id", taskId).
			WithCtx(ctx)
	}
	return task.State == sdktask.TaskStateSuccess, nil
}

func (p *AssetProcedure) Retry(ctx context.Context, record *biz.ProcedureRecord) error {
	return p.retryHelper.retry(
		ctx,
		record,
		p.PollResult,
		p.executeForRetry,
		p.OnSuccess,
		p.OnFailure,
	)
}

func (p *AssetProcedure) executeForRetry(ctx context.Context, note *model.Note) (string, error) {
	processor := assetprocess.NewProcessor(note.Type, p.bizz)
	taskId, err := processor.Process(ctx, note)
	if err != nil {
		xlog.Msg("asset procedure retry process failed").
			Err(err).
			Extra("note_id", note.NoteId).
			Errorx(ctx)
		return "", err
	}
	return taskId, nil
}
