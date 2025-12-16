package procedure

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

var _ Procedure = (*AuditProcedure)(nil)

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
func (p *AuditProcedure) PreStart(ctx context.Context, note *model.Note) (bool, error) {
	err := p.noteCreatorBiz.TransferNoteStateToAuditing(ctx, note.NoteId)
	if err != nil {
		return false, xerror.Wrapf(err, "audit procedure set note state auditing failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	// TODO 后续接入审核系统时返回 true
	return false, nil
}

func (p *AuditProcedure) Execute(ctx context.Context, note *model.Note) (string, error) {
	// TODO 调用审核系统
	return "", nil
}

func (p *AuditProcedure) OnSuccess(ctx context.Context, result *ProcedureResult) (bool, error) {
	// TODO 审核通过 设置状态为 AuditPassed
	err := p.noteCreatorBiz.TransferNoteStateToAuditPassed(ctx, result.NoteId)
	if err != nil {
		return false, xerror.Wrapf(err, "audit procedure set note state audit passed failed").
			WithExtra("note_id", result.NoteId).
			WithCtx(ctx)
	}
	return true, nil
}

func (p *AuditProcedure) OnFailure(ctx context.Context, result *ProcedureResult) (bool, error) {
	// TODO 审核不通过 设置状态为 Rejected
	err := p.noteCreatorBiz.TransferNoteStateToRejected(ctx, result.NoteId)
	if err != nil {
		return false, xerror.Wrapf(err, "audit procedure set note state rejected failed").
			WithExtra("note_id", result.NoteId).
			WithCtx(ctx)
	}
	return true, nil
}

func (p *AuditProcedure) PollResult(ctx context.Context, taskId string) (PollState, any, error) {
	// TODO 轮询审核结果
	return PollStateSuccess, nil, nil
}

func (p *AuditProcedure) Retry(ctx context.Context, record *biz.ProcedureRecord) error {
	// TODO 重试审核
	return nil
}

// TODO audit未实现 这里先自动成功
var _ AutoCompleter = (*AuditProcedure)(nil)

func (p *AuditProcedure) AutoComplete(
	ctx context.Context,
	note *model.Note,
	taskId string,
) (success, autoComplete bool, arg any) {
	return true, true, nil
}
