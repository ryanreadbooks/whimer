package procedure

import (
	"context"

	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

var _ Procedure = (*PublishProcedure)(nil)

// 笔记发布
// 
// 流水线中的最后一步
type PublishProcedure struct {
}

func NewPublishProcedure() *PublishProcedure {
	return &PublishProcedure{}
}

func (p *PublishProcedure) Type() model.ProcedureType {
	return model.ProcedureTypePublish
}

func (p *PublishProcedure) PreStart(ctx context.Context, note *model.Note) error {
	
	return nil
}

func (p *PublishProcedure) Execute(ctx context.Context, note *model.Note) (string, error) {
	return "", nil
}

func (p *PublishProcedure) OnSuccess(ctx context.Context, noteId int64, taskId string) error {
	return nil
}

func (p *PublishProcedure) OnFailure(ctx context.Context, noteId int64, taskId string) error {
	return nil
}

func (p *PublishProcedure) PollResult(ctx context.Context, taskId string) (bool, error) {
	return false, nil
}

func (p *PublishProcedure) Retry(ctx context.Context, record *biz.ProcedureRecord) error {
	return nil
}
