package messaging

import (
	"context"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/search/internal/srv"
)

func startHandlingNoteEvents(svc *srv.Service) {
	ctx := context.Background()
	concurrent.SafeGo(func() {
		xlog.Msg("start handling note events").Info()
		for {
			// TODO 优化为批量读入FetchMessage再写入es 达到一定数量或者到了一定时间触发写入es
			m, err := noteEventReader.FetchMessage(ctx)
			if err != nil {
				xlog.Msg("when handling note events, fetch message failed").Err(err).Error()
				break
			}

			// we got message
			ctx := xkafka.ContextFromKafkaHeaders(m.Headers)
			// start handling
			err = svc.DocumentSrv.DispatchNoteEvent(ctx, &m)
			if err != nil {
				xlog.Msg("handle note event failed").Err(err).Errorx(ctx)
			}
			err = noteEventReader.CommitMessages(ctx, m)
			if err != nil {
				xlog.Msg("handle note commit messages failed").Err(err).Errorx(ctx)
			}
		}
	})
}
