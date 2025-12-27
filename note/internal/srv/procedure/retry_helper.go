package procedure

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

type retryExecuteFunc func(ctx context.Context, note *model.Note) (taskId string, err error)
type retryPollResultFunc func(ctx context.Context, taskId string) (PollState, any, error)
type onCompleteFunc func(ctx context.Context, result *ProcedureResult) (bool, error)

// procedure的通用重试逻辑
type retryHelper struct {
	txHelper         *txHelper
	noteBiz          *biz.NoteBiz
	noteProcedureBiz *biz.NoteProcedureBiz
}

func newRetryHelper(bizz *biz.Biz) *retryHelper {
	return &retryHelper{
		txHelper:         newTxHelper(bizz),
		noteBiz:          bizz.Note,
		noteProcedureBiz: bizz.Procedure,
	}
}

func (h *retryHelper) retry(
	ctx context.Context,
	record *biz.ProcedureRecord,
	pollFn retryPollResultFunc,
	execFn retryExecuteFunc,
	onSuccess, onFailure onCompleteFunc,
) error {
	// 存在 taskId 表明注册成功，主动轮询结果
	if record.TaskId != "" {
		return h.pollAndComplete(ctx, record, pollFn, onSuccess, onFailure)
	}

	// 检查是否允许重试
	// 超过重试视为失败
	if record.CurRetry >= record.MaxRetryCnt {
		xlog.Msg("retry helper retry exhausted, mark as failed").
			Extras("record_id", record.Id, "note_id", record.NoteId, "protype", record.Protype).
			Infox(ctx)
		return h.txHelper.txHandleFailure(ctx, &CompleteResult{
			NoteId:  record.NoteId,
			Protype: record.Protype,
			TaskId:  record.TaskId,
			Success: false,
			Arg:     nil,
		}, onFailure)
	}

	// 不存在 taskId，重新执行任务
	return h.reExecute(ctx, record, execFn, onFailure)
}

func (h *retryHelper) pollAndComplete(
	ctx context.Context,
	record *biz.ProcedureRecord,
	pollFn retryPollResultFunc,
	onSuccess, onFailure onCompleteFunc,
) error {
	state, arg, err := pollFn(ctx, record.TaskId)
	if err != nil {
		xlog.Msg("retry helper poll result failed").
			Err(err).
			Extras("record_id", record.Id, "task_id", record.TaskId).
			Errorx(ctx)
		return err
	}

	result := &CompleteResult{
		NoteId:  record.NoteId,
		Protype: record.Protype,
		TaskId:  record.TaskId,
		Success: state == PollStateSuccess,
		Arg:     arg,
	}

	switch state {
	case PollStateSuccess:
		return h.txHelper.txHandleSuccess(ctx, result, onSuccess)
	case PollStateFailure:
		return h.txHelper.txHandleFailure(ctx, result, onFailure)
	default:
		// PollStateRunning: 任务仍在运行中，等待下次轮询
		xlog.Msg("retry helper task still running, skip").
			Extras("record_id", record.Id, "task_id", record.TaskId).
			Infox(ctx)
		return nil
	}
}

func (h *retryHelper) reExecute(
	ctx context.Context,
	record *biz.ProcedureRecord,
	execFn retryExecuteFunc,
	onFailure onCompleteFunc,
) error {
	// 获取笔记原始信息
	note, err := h.noteBiz.GetNoteWithoutCache(ctx, record.NoteId)
	if err != nil {
		xlog.Msg("retry helper get note failed").
			Err(err).
			Extras("record_id", record.Id, "note_id", record.NoteId).
			Errorx(ctx)
		return err
	}

	// 计算下次检查时间
	retryInterval := config.Conf.RetryConfig.ProcedureRetry.TaskRegister.RetryInterval
	nextCheckTime := time.Now().Add(retryInterval)

	// 执行任务
	taskId, err := execFn(ctx, note)
	if err != nil {
		return h.handleExecuteFailure(ctx, record, nextCheckTime, onFailure)
	}

	return h.handleExecuteSuccess(ctx, record, taskId, nextCheckTime)
}

// 远程任务还是出错，则更新重试计数
func (h *retryHelper) handleExecuteFailure(
	ctx context.Context,
	record *biz.ProcedureRecord,
	nextCheckTime time.Time,
	onFailure onCompleteFunc,
) error {
	nowRetryCnt := record.CurRetry + 1
	shouldMarkFailure := nowRetryCnt >= record.MaxRetryCnt

	// 更新重试计数
	if err := h.noteProcedureBiz.UpdateRetry(
		ctx,
		record.NoteId,
		record.Protype,
		nextCheckTime.Unix(),
		shouldMarkFailure,
	); err != nil {
		xlog.Msg("retry helper update retry failed").
			Err(err).
			Extras("record_id", record.Id).
			Errorx(ctx)
	}

	// 最后一次重试仍失败，标记流程失败
	if shouldMarkFailure {
		return h.txHelper.txHandleFailure(ctx, &CompleteResult{
			NoteId:  record.NoteId,
			Protype: record.Protype,
			TaskId:  record.TaskId,
			Success: false,
			Arg:     nil,
		}, onFailure)
	}

	return nil
}

// 远程任务执行成功
func (h *retryHelper) handleExecuteSuccess(
	ctx context.Context,
	record *biz.ProcedureRecord,
	taskId string,
	nextCheckTime time.Time,
) error {
	record.TaskId = taskId
	record.NextCheckTime = nextCheckTime.Unix()
	record.CurRetry++

	if err := h.noteProcedureBiz.UpdateTaskIdRetryNextCheckTime(ctx, record); err != nil {
		xlog.Msg("retry helper update record failed").
			Err(err).
			Extras("record_id", record.Id).
			Errorx(ctx)
	}

	return nil
}
