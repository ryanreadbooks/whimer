package events

import (
	"context"
	"encoding/json"

	"github.com/ryanreadbooks/whimer/asset-job/internal/model"
	"github.com/ryanreadbooks/whimer/asset-job/internal/srv"
	"github.com/ryanreadbooks/whimer/misc/xkq"
	"github.com/ryanreadbooks/whimer/misc/xlog"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/queue"
)

func registerNoteImageEvent(c kq.KqConf, svc *srv.Service) queue.MessageQueue {
	return kq.MustNewQueue(c, noteImageEventConsumer(svc))
}

// 图片上传成功的处理动作
func noteImageEventConsumer(svc *srv.Service) xkq.Consumer {
	return func(ctx context.Context, key, value string) error {
		var event model.MinioEvent
		err := json.Unmarshal([]byte(value), &event)
		if err != nil {
			xlog.Msg("note image uploaded event handler unable to unmarshal value").Err(err).Errorx(ctx)
			return err
		}

		return svc.NoteImageService.OnImageUploaded(ctx, &event)
	}
}
