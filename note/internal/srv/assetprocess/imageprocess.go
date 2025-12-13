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
	baseProcessor

	bizz *biz.Biz
}

func newImageProcessor(biz *biz.Biz) Processor {
	return &ImageProcessor{bizz: biz}
}

// 处理笔记图片
func (p *ImageProcessor) Process(ctx context.Context, note *model.Note) (string, error) {
	// 注册任务+回调
	callbackUrl := encodeCallbackUrl(config.Conf.DevCallbacks.NoteProcessCallback, note.NoteId)
	taskId, err := dep.GetConductProducer().Schedule(
		ctx,
		global.NoteImageProcessTaskType,
		note,
		conductor.ScheduleOptions{
			Namespace:   global.NoteProcessNamespace,
			CallbackUrl: callbackUrl,
			MaxRetry:    5,
			ExpireAfter: 1 * time.Hour,
		})
	if err != nil {
		return "", xerror.Wrapf(err, "srv creator schedule task failed").
			WithExtra("note_id", note.NoteId).
			WithCtx(ctx)
	}

	return taskId, nil
}
