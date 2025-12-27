package procedure

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

var _ Procedure = (*PublishProcedure)(nil)

// 笔记发布
//
// 流水线中的最后一步
type PublishProcedure struct {
	noteBiz        *biz.NoteBiz
	noteCreatorBiz *biz.NoteCreatorBiz
}

func NewPublishProcedure(bizz *biz.Biz) *PublishProcedure {
	return &PublishProcedure{
		noteBiz:        bizz.Note,
		noteCreatorBiz: bizz.Creator,
	}
}

func (p *PublishProcedure) Type() model.ProcedureType {
	return model.ProcedureTypePublish
}

// 发布流程
func (p *PublishProcedure) PreStart(ctx context.Context, note *model.Note) (bool, error) {
	err := p.noteCreatorBiz.TransferNoteStateToPublished(ctx, note)
	if err != nil {
		return false, xerror.Wrapf(err, "publish procedure set note state published failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	return false, nil
}

// 广播笔记发布事件
func (p *PublishProcedure) Execute(ctx context.Context, note *model.Note) (string, error) {
	if note.Privacy == model.PrivacyPrivate {
		return "", nil
	}

	return "", nil
}

func (p *PublishProcedure) OnSuccess(ctx context.Context, result *ProcedureResult) (bool, error) {
	return false, nil
}

func (p *PublishProcedure) OnFailure(ctx context.Context, result *ProcedureResult) (bool, error) {
	return false, nil
}

func (p *PublishProcedure) PollResult(ctx context.Context, taskId string) (PollState, any, error) {
	return PollStateSuccess, nil, nil
}

func (p *PublishProcedure) Retry(ctx context.Context, record *biz.ProcedureRecord) error {
	return nil
}
