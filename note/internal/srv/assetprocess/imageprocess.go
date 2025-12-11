package assetprocess

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/note/internal/biz"
	"github.com/ryanreadbooks/whimer/note/internal/config"
	"github.com/ryanreadbooks/whimer/note/internal/global"
	"github.com/ryanreadbooks/whimer/note/internal/infra/dep"
	"github.com/ryanreadbooks/whimer/note/internal/model"

	conductor "github.com/ryanreadbooks/whimer/conductor/pkg/sdk/producer"
	"github.com/ryanreadbooks/whimer/misc/xerror"
)

type ImageProcessor struct {
	bizz biz.Biz
}

func newImageProcessor(biz biz.Biz) Processor {
	return &ImageProcessor{bizz: biz}
}

// 处理笔记图片
func (p *ImageProcessor) Process(ctx context.Context, note *model.Note) error {
	// 先设置状态
	err := p.bizz.Creator.SetNoteStateProcessing(ctx, note.NoteId)
	if err != nil {
		return xerror.Wrapf(err, "srv creator set note state processing failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	// 注册任务+回调
	taskId, err := dep.GetConductProducer().Schedule(
		ctx,
		global.NoteImageProcessTaskType,
		note,
		conductor.ScheduleOptions{
			Namespace:   global.NoteProcessNamespace,
			CallbackUrl: config.Conf.DevCallbacks.NoteProcessCallback,
			MaxRetry:    5,
			ExpireAfter: 1 * time.Hour,
		})
	if err != nil {
		return xerror.Wrapf(err, "srv creator schedule task failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	_, err = p.bizz.Process.CreateRecord(ctx, note.NoteId, taskId)
	if err != nil {
		// 实际注册任务成功了 但是本地落盘失败
		// TODO 打点上报
		return xerror.Wrapf(err, "srv creator create process record failed").
			WithExtra("note_id", note.NoteId).
			WithExtra("task_id", taskId).
			WithCtx(ctx)
	}

	// TODO 后台主动轮询taskId查询状态

	return nil
}
