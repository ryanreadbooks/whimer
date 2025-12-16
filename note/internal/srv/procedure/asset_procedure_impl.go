package procedure

import (
	"context"
	"encoding/json"
	"errors"

	conductor "github.com/ryanreadbooks/whimer/conductor/pkg/sdk/producer"
	sdktask "github.com/ryanreadbooks/whimer/conductor/pkg/sdk/task"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/global"
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
		retryHelper:       newRetryHelper(bizz),
	}
}

func (p *AssetProcedure) Type() model.ProcedureType {
	return model.ProcedureTypeAssetProcess
}

func (p *AssetProcedure) PreStart(ctx context.Context, note *model.Note) (bool, error) {
	err := p.noteCreatorBiz.TransferNoteStateToProcessing(ctx, note.NoteId)
	if err != nil {
		return false, xerror.Wrapf(err, "asset procedure set note state processing failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}
	return true, nil
}

func (p *AssetProcedure) Execute(ctx context.Context, note *model.Note) (string, error) {
	if note.Type == model.AssetTypeVideo {
		processor := assetprocess.NewProcessor(note.Type, p.bizz)
		taskId, err := processor.Process(ctx, note)
		if err != nil {
			return "", xerror.Wrapf(err, "asset procedure execute failed").
				WithExtra("note_id", note.NoteId).
				WithCtx(ctx)
		}
		return taskId, nil
	}

	// 图片资源不处理
	return "", nil
}

// 返回的note只有basic信息 没有额外信息
func (p *AssetProcedure) upgradeStateCheck(
	ctx context.Context,
	noteId int64,
	state model.NoteState,
) (note *model.Note, abort bool) {
	note, err := p.noteBiz.GetNoteWithoutCache(ctx, noteId)
	if err != nil {
		xlog.Msg("asset procedure get note failed, try to update state without checking").
			Err(err).
			Extras("noteId", noteId).
			Errorx(ctx)
	}

	if errors.Is(err, global.ErrNoteNotFound) {
		return nil, true
	}

	if note != nil && note.State > state {
		// 已经有了最新状态了 可能已经被更新
		return nil, true
	}

	return note, false
}

func (p *AssetProcedure) OnSuccess(ctx context.Context, result *ProcedureResult) (bool, error) {
	noteId, taskId, arg := result.NoteId, result.TaskId, result.Arg

	// 简单幂等保证
	note, abort := p.upgradeStateCheck(ctx, noteId, model.NoteStateProcessed)
	if abort {
		return true, nil
	}

	if err := p.noteCreatorBiz.TransferNoteStateToProcessed(ctx, noteId); err != nil {
		return false, xerror.Wrapf(err, "asset procedure set note state processed failed").
			WithExtra("noteId", noteId).
			WithCtx(ctx)
	}

	var state = make(map[string]bool)

	// 如果视频资源此时需要更新视频资源的metadata
	if note != nil && note.Type == model.AssetTypeVideo {
		metadata, ok := arg.([]*model.VideoAssetMetadata)
		if ok {
			metaMap := make(map[string][]byte)
			state["meta_type_asset"] = true
			for _, meta := range metadata {
				metaBytes, err := json.Marshal(meta.Info)
				if err != nil {
					xlog.Msg("asset procedure on success failed to marshal video asset metadata").
						Err(err).
						Extra("noteId", noteId).
						Extra("meta", meta).
						Errorx(ctx)
					continue
				}
				metaMap[meta.Key] = metaBytes
			}

			// 更新video asset metadata
			err := p.noteCreatorBiz.BatchUpdateAssetMeta(ctx, noteId, metaMap)
			if err != nil {
				// 次要信息 可以仅打印日志
				xlog.Msg("asset procedure on success failed to batch update video asset metadata").
					Err(err).
					Extra("noteId", noteId).
					Errorx(ctx)
			} else {
				state["meta_updated"] = true
			}
		}
	}

	xlog.Msg("asset procedure on success completed").
		Extras("taskId", taskId, "noteId", noteId, "state", state).
		Infox(ctx)

	return true, nil
}

func (p *AssetProcedure) OnFailure(ctx context.Context, result *ProcedureResult) (bool, error) {
	noteId, taskId := result.NoteId, result.TaskId

	// 简单幂等保证
	_, abort := p.upgradeStateCheck(ctx, noteId, model.NoteStateProcessFailed)
	if abort {
		return true, nil
	}

	if err := p.noteCreatorBiz.TransferNoteStateToProcessFailed(ctx, noteId); err != nil {
		return false, xerror.Wrapf(err, "asset procedure set note state failed failed").
			WithExtra("noteId", noteId).
			WithCtx(ctx)
	}

	xlog.Msg("asset procedure on failure completed").
		Extras("taskId", taskId, "noteId", noteId).
		Infox(ctx)

	return true, nil
}

func (p *AssetProcedure) PollResult(ctx context.Context, taskId string) (PollState, any, error) {
	task, err := p.conductorProducer.GetTask(ctx, taskId)
	if err != nil {
		return PollStateRunning, nil, xerror.Wrapf(err, "asset procedure poll result failed").
			WithExtra("task_id", taskId).
			WithCtx(ctx)
	}
	switch task.State {
	case sdktask.TaskStateSuccess:
		return PollStateSuccess, task.OutputArgs, nil
	case sdktask.TaskStateFailure:
		return PollStateFailure, task.OutputArgs, nil
	default:
		return PollStateRunning, nil, nil
	}
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

var _ AutoCompleter = (*AssetProcedure)(nil)

func (p *AssetProcedure) AutoComplete(
	ctx context.Context,
	note *model.Note,
	taskId string,
) (success, autoComplete bool, arg any) {
	if note.Type == model.AssetTypeImage {
		return true, true, nil
	}

	return false, false, nil
}
