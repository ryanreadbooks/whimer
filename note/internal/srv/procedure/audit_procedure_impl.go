package procedure

import (
	"context"
	"errors"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

var (
	_ Procedure     = (*AuditProcedure)(nil)
	_ AutoCompleter = (*AuditProcedure)(nil) // TODO audit未实现 这里先自动成功
)

// 笔记审核
//
// 流水线中资源处理完成后进入审核流程
type AuditProcedure struct {
	noteBiz        *biz.NoteBiz
	noteCreatorBiz *biz.NoteCreatorBiz
}

func NewAuditProcedure(bizz *biz.Biz) *AuditProcedure {
	return &AuditProcedure{
		noteBiz:        bizz.Note,
		noteCreatorBiz: bizz.Creator,
	}
}

func (p *AuditProcedure) Type() model.ProcedureType {
	return model.ProcedureTypeAudit
}

// 审核流程初始化
func (p *AuditProcedure) BeforeExecute(ctx context.Context, note *model.Note) (bool, error) {
	err := p.noteCreatorBiz.TransferNoteStateToAuditing(ctx, note)
	if err != nil {
		return false, xerror.Wrapf(err, "audit procedure set note state auditing failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	// TODO 后续接入审核系统时返回 true
	return false, nil
}

// 返回的note只有basic信息 没有额外信息
func (p *AuditProcedure) upgradeStateCheck(
	ctx context.Context,
	noteId int64,
	compareState model.NoteState,
) (note *model.Note, abort bool) {
	note, err := p.noteBiz.GetNoteCoreWithoutCache(ctx, noteId)
	if err != nil {
		xlog.Msg("audit procedure get note failed, try to update state without checking").
			Err(err).
			Extras("noteId", noteId).
			Errorx(ctx)
	}

	if errors.Is(err, global.ErrNoteNotFound) {
		return nil, true
	}

	if note != nil && (note.State == model.NoteStateAuditPassed || note.State == model.NoteStateRejected) {
		return nil, true
	}

	if note != nil && note.State > compareState {
		// 已经有了最新状态了 可能已经被更新
		return nil, true
	}

	return note, false
}

func (p *AuditProcedure) Execute(ctx context.Context, note *model.Note) (string, error) {
	// TODO 调用审核系统
	return "", nil
}

func (p *AuditProcedure) OnSuccess(ctx context.Context, result *ProcedureResult) (bool, error) {
	// 审核通过 设置状态为 AuditPassed
	note, abort := p.upgradeStateCheck(ctx, result.NoteId, model.NoteStateAuditPassed)
	if abort {
		return true, nil
	}

	if note == nil {
		note = &model.Note{
			NoteId: result.NoteId,
		}
	}

	err := p.noteCreatorBiz.TransferNoteStateToAuditPassed(ctx, note)
	if err != nil {
		return false, xerror.Wrapf(err, "audit procedure set note state audit passed failed").
			WithExtra("note_id", result.NoteId).
			WithCtx(ctx)
	}
	return true, nil
}

// TODO 失败是审核失败 不是审核不通过
func (p *AuditProcedure) OnFailure(ctx context.Context, result *ProcedureResult) (bool, error) {
	// 审核不通过 设置状态为 Rejected
	note, abort := p.upgradeStateCheck(ctx, result.NoteId, model.NoteStateRejected)
	if abort {
		return true, nil
	}

	if note == nil {
		note = &model.Note{
			NoteId: result.NoteId,
		}
	}

	err := p.noteCreatorBiz.TransferNoteStateToRejected(ctx, note)
	if err != nil {
		return false, xerror.Wrapf(err, "audit procedure set note state rejected failed").
			WithExtra("note_id", result.NoteId).
			WithCtx(ctx)
	}
	return true, nil
}

func (p *AuditProcedure) ObAbort(ctx context.Context, note *model.Note, taskId string) error {
	return nil
}

func (p *AuditProcedure) PollResult(ctx context.Context, record *biz.ProcedureRecord) (PollState, any, error) {
	// TODO 轮询审核结果
	return PollStateSuccess, nil, nil
}

func (p *AuditProcedure) Retry(ctx context.Context, record *biz.ProcedureRecord) (string, error) {
	// TODO 重试审核
	return "", nil
}

func (p *AuditProcedure) AutoComplete(
	ctx context.Context,
	note *model.Note,
	taskId string,
) (success, autoComplete bool, arg any) {
	return true, true, nil
}
