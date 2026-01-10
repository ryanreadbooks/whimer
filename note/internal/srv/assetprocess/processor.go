package assetprocess

import (
	"context"

	sdktask "github.com/ryanreadbooks/whimer/conductor/pkg/sdk/task"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/note/internal/model"
)

type Processor interface {
	Process(ctx context.Context, note *model.Note) (string, error)
	GetTaskResult(ctx context.Context, taskId string) (output []byte, success bool, err error)
}

func NewProcessor(noteType model.NoteType, biz *biz.Biz) Processor {
	switch noteType {
	case model.AssetTypeImage:
		return newImageProcessor(biz)
	case model.AssetTypeVideo:
		return newVideoProcessor(biz)
	}
	return nil
}

type baseProcessor struct {
}

func (p *baseProcessor) GetTaskResult(
	ctx context.Context,
	taskId string) (
	output []byte, success bool, err error,
) {
	task, err := dep.GetConductProducer().GetTask(ctx, taskId)
	if err != nil {
		err = xerror.Wrapf(err, "srv creator get task result failed").
			WithExtra("task_id", taskId).
			WithCtx(ctx)
		return
	}

	success = task.State == sdktask.TaskStateSuccess
	output = task.OutputArgs

	return
}
