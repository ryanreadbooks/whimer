package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xerror"
	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xretry"
	"github.com/ryanreadbooks/whimer/pilot/internal/app"

	noteevent "github.com/ryanreadbooks/whimer/note/pkg/event/note"
)

func startNoteEventConsumer(manager *app.Manager) {
	backoff := xretry.NewPermenantBackoff(time.Millisecond*100, time.Second*8, 2.0)
	concurrent.SafeGo2(rootCtx, concurrent.SafeGo2Opt{
		Name: "pilot.note_event.consumer",
		Job: func(ctx context.Context) error {
			xlog.Msg("start consuming note event")

			trapped := false
			for {
				kMsg, err := noteEventConsumer.ReadMessage(ctx)
				if err != nil {
					// 区分不同类型的错误
					if errors.Is(err, context.Canceled) {
						// 上下文取消才认为是退出
						xlog.Msg("pilot note.event.consumer context done").Err(err).Info()
						break
					}

					xlog.Msg("pilot note.event.consumer read message err, will retry").Err(err).Errorx(ctx)

					rest, _ := backoff.NextBackOff()
					trapped = true
					select {
					case <-time.After(rest):
					case <-ctx.Done():
						// check again
						xlog.Msg("pilot note.event.consumer context done during retry delay").Info()
						break
					}
					continue
				}

				if trapped {
					backoff.Success()
					trapped = false
				}

				var ev noteevent.NoteEvent
				err = json.Unmarshal(kMsg.Value, &ev)
				if err != nil {
					xlog.Msg("pilot note.event.consumer json.Unmarshal err").Err(err).Errorx(ctx)
					continue
				}

				msgCtx := xkafka.ContextFromKafkaHeaders(kMsg.Headers)
				if err := handleNoteEvent(msgCtx, manager, ev); err != nil {
					xlog.Msg("pilot note.event.consumer handle note event err").Err(err).Errorx(ctx)
				}
			}

			return nil
		},
	})
}

func handleNoteEvent(ctx context.Context, manager *app.Manager, ev noteevent.NoteEvent) error {
	switch ev.Type {
	case noteevent.NotePublished:
		// 解码payload
		var data noteevent.NotePublishedEventData
		if err := json.Unmarshal(ev.Payload, &data); err != nil {
			return xerror.Wrapf(err, "pilot note.event.consumer handle note published event failed to unmarshal payload").
				WithExtra("note_id", ev.NoteId).
				WithCtx(ctx)
		}

		err := manager.NoteEventApp.OnNotePublished(ctx, data)
		if err != nil {
			return xerror.Wrapf(err, "pilot note.event.consumer handle note published event failed to add note to search").
				WithExtra("note_id", ev.NoteId).
				WithCtx(ctx)
		}
	case noteevent.NoteDeleted:
		var data noteevent.NoteDeletedEventData
		if err := json.Unmarshal(ev.Payload, &data); err != nil {
			return xerror.Wrapf(err, "pilot note.event.consumer handle note deleted event failed to unmarshal payload").
				WithExtra("note_id", ev.NoteId).
				WithCtx(ctx)
		}

		err := manager.NoteEventApp.OnNoteDeleted(ctx, data)
		if err != nil {
			return xerror.Wrapf(err, "pilot note.event.consumer handle note deleted event failed to delete note from search").
				WithExtra("note_id", ev.NoteId).
				WithCtx(ctx)
		}

	default:
		// TODO 对其它类型不处理
	}

	return nil
}
