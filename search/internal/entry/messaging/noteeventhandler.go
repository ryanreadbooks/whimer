package messaging

import (
	"context"

	"github.com/ryanreadbooks/whimer/search/internal/srv"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xlog"
)

func startHandlingNoteEvents(svc *srv.Service) {
	ctx := context.Background()
	concurrent.SafeGo(func() {
		xlog.Msg("start handling note events").Info()
		for {
			msgs, err := noteEventBatchReader.BatchFetchMessages(ctx)
			if err != nil {
				xlog.Msg("when handling note events, fetch message failed").Err(err).Error()
				break
			}

			err = svc.DocumentSrv.DispatchNoteEvents(ctx, msgs)
			if err != nil {
				xlog.Msg("handle note events failed").Err(err).Errorx(ctx)
			}
			err = noteEventBatchReader.CommitMessages(ctx, msgs...)
			if err != nil {
				xlog.Msg("handle note commit messages failed").Err(err).Errorx(ctx)
			}
		}
	})
}
