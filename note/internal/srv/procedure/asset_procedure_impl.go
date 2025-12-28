package procedure

import (
	"bytes"
	"compress/zlib"
	"context"
	"encoding/json"
	"errors"
	"io"

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

var (
	_ Procedure              = (*AssetProcedure)(nil)
	_ AutoCompleter          = (*AssetProcedure)(nil)
	_ ProcedureParamProvider = (*AssetProcedure)(nil)
)

// AssetProcedure 资源处理流程 负责笔记资源（图片、视频等）的处理
type AssetProcedure struct {
	bizz           *biz.Biz
	noteBiz        *biz.NoteBiz
	noteCreatorBiz *biz.NoteCreatorBiz

	conductorProducer *conductor.Client

	// 重试通用逻辑
	retryHelper *retryHelper
}

func NewAssetProcedure(bizz *biz.Biz) *AssetProcedure {
	return &AssetProcedure{
		bizz:              bizz,
		noteBiz:           bizz.Note,
		noteCreatorBiz:    bizz.Creator,
		conductorProducer: dep.GetConductProducer(),
		retryHelper:       newRetryHelper(bizz),
	}
}

func (p *AssetProcedure) Type() model.ProcedureType {
	return model.ProcedureTypeAssetProcess
}

func (p *AssetProcedure) PreStart(ctx context.Context, note *model.Note) (bool, error) {
	// 如果是笔记更新场景 不需要更新资源key的情况下不需要重走资源流程
	if note.State == model.NoteStateInit {
		err := p.noteCreatorBiz.TransferNoteStateToProcessing(ctx, note)
		if err != nil {
			return false, xerror.Wrapf(err, "asset procedure set note state processing failed").
				WithExtra("note_id", note.NoteId).
				WithCtx(ctx)
		}
		return true, nil
	}

	return false, nil
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
	compareState model.NoteState,
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

	if note != nil && note.State > compareState {
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

	if note == nil {
		note = &model.Note{
			NoteId: noteId,
		}
	}
	if err := p.noteCreatorBiz.TransferNoteStateToProcessed(ctx, note); err != nil {
		return false, xerror.Wrapf(err, "asset procedure set note state processed failed").
			WithExtra("noteId", noteId).
			WithCtx(ctx)
	}

	state := make(map[string]bool)

	// 如果视频资源此时需要更新视频资源的metadata
	if note != nil && note.Type == model.AssetTypeVideo {
		metadata, ok := arg.([]*model.VideoAsset)
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
	note, abort := p.upgradeStateCheck(ctx, noteId, model.NoteStateProcessFailed)
	if abort {
		return true, nil
	}

	if note == nil {
		note = &model.Note{
			NoteId: noteId,
		}
	}
	if err := p.noteCreatorBiz.TransferNoteStateToProcessFailed(ctx, note); err != nil {
		return false, xerror.Wrapf(err, "asset procedure set note state failed failed").
			WithExtra("noteId", noteId).
			WithCtx(ctx)
	}

	xlog.Msg("asset procedure on failure completed").
		Extras("taskId", taskId, "noteId", noteId).
		Infox(ctx)

	return true, nil
}

func (p *AssetProcedure) Abort(ctx context.Context, note *model.Note, taskId string) error {
	if note.Type != model.AssetTypeVideo {
		return nil
	}

	err := p.conductorProducer.AbortTask(ctx, taskId)
	if err != nil {
		return xerror.Wrapf(err, "asset procedure abort task failed").
			WithExtra("task_id", taskId).
			WithCtx(ctx)
	}

	return nil
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

// 受到重试调度 重新执行远程任务
func (p *AssetProcedure) executeForRetry(ctx context.Context, note *model.Note, params []byte) (string, error) {
	processor := assetprocess.NewProcessor(note.Type, p.bizz)
	if note.Videos == nil {
		var err error
		note.Videos, err = p.deserializeVideoParam(params)
		if err != nil {
			// 无法重建videos请求资源 直接失败
			return "", xerror.Wrapf(err, "asset procedure retry deserialize video param failed").
				WithExtra("note_id", note.NoteId).
				WithCtx(ctx)
		}
	}

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

func (p *AssetProcedure) Provide(note *model.Note) []byte {
	if note.Type == model.AssetTypeVideo {
		// 保存原始视频的参数 比如key等
		return p.serializeVideoParam(note.Videos)
	}

	return nil
}

type assetVideoParam struct {
	RawUrl    string                 `json:"raw_url"`
	RawBucket string                 `json:"raw_bucket"`
	Items     []*assetVideoParamItem `json:"items"`
}

type assetVideoParamItem struct {
	Key    string `json:"key"`
	Bucket string `json:"bucket"`
}

func (p *AssetProcedure) serializeVideoParam(video *model.NoteVideo) []byte {
	param := &assetVideoParam{
		RawUrl:    video.GetRawUrl(),
		RawBucket: video.GetRawBucket(),
	}
	for _, item := range video.Items {
		param.Items = append(param.Items, &assetVideoParamItem{
			Key:    item.Key,
			Bucket: item.GetBucket(),
		})
	}
	paramBytes, err := json.Marshal(param)
	if err != nil {
		return nil
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(paramBytes)))
	zw := zlib.NewWriter(buf)
	if _, err = zw.Write(paramBytes); err != nil {
		xlog.Msg("asset procedure serialize video param failed to write").Err(err).Error()
		return nil
	}
	if err = zw.Close(); err != nil {
		xlog.Msg("asset procedure serialize video param failed to close").Err(err).Error()
		return nil
	}
	return buf.Bytes()
}

func (p *AssetProcedure) deserializeVideoParam(input []byte) (*model.NoteVideo, error) {
	gr, err := zlib.NewReader(bytes.NewReader(input))
	if err != nil {
		return nil, xerror.Wrap(err)
	}
	defer gr.Close()
	paramBytes, err := io.ReadAll(gr)
	if err != nil {
		return nil, xerror.Wrap(err)
	}

	var param assetVideoParam
	err = json.Unmarshal(paramBytes, &param)
	if err != nil {
		return nil, xerror.Wrap(err)
	}
	video := &model.NoteVideo{}
	video.SetRawUrl(param.RawUrl)
	video.SetRawBucket(param.RawBucket)
	for _, item := range param.Items {
		vi := &model.NoteVideoItem{
			Key: item.Key,
		}
		vi.SetBucket(item.Bucket)
		video.Items = append(video.Items, vi)
	}

	return video, nil
}
