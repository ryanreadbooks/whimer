package events

import (
	"github.com/ryanreadbooks/whimer/asset-job/internal/config"
	"github.com/ryanreadbooks/whimer/asset-job/internal/srv"
	"github.com/zeromicro/go-zero/core/queue"
)

func Init(c *config.Config, svc *srv.Service) []queue.MessageQueue {
	var qs []queue.MessageQueue
	qs = append(qs, regNoteImageUploadedEventConsumer(c.NoteAssetEventKafka.AsKqConf(), svc))

	return qs
}
