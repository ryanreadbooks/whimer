package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	xkafka "github.com/ryanreadbooks/whimer/misc/xkq/kafka"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/misc/xretry"
	"github.com/ryanreadbooks/whimer/pilot/internal/biz"
	"github.com/ryanreadbooks/whimer/pilot/internal/infra/dao/kafka/sysmsg"
)

// 系统消息懒删除事件消费
//
// Topic: pilot_sysmsg_deletion_topic
func startSysMsgDeletionConsumer(bizz *biz.Biz) {
	backoff := xretry.NewPermenantBackoff(time.Millisecond*100, time.Second*8, 2.0)
	concurrent.SafeGo2(rootCtx, concurrent.SafeGo2Opt{
		Name: "pilot.sysmsg.deletion.consumer",
		Job: func(ctx context.Context) error {
			xlog.Msg("start consuming sysmsg.deletion")

			trapped := false
			for {
				kMsg, err := sysMsgDeletionConsumer.ReadMessage(ctx)
				if err != nil {
					// 区分不同类型的错误
					if errors.Is(err, context.Canceled) {
						// 上下文取消才认为是退出
						xlog.Msg("pilot sysmsg.deletion.consumer context done").Err(err).Info()
						break
					}

					xlog.Msg("pilot sysmsg.deletion.consumer read message err, will retry").Err(err).Errorx(ctx)

					rest, _ := backoff.NextBackOff()
					trapped = true
					select {
					case <-time.After(rest):
					case <-ctx.Done():
						// check again
						xlog.Msg("pilot sysmsg.deletion.consumer context done during retry delay").Info()
						break
					}
					continue
				}

				if trapped {
					backoff.Success()
					trapped = false
				}

				var ev sysmsg.DeletionEvent
				err = json.Unmarshal(kMsg.Value, &ev)
				if err != nil {
					xlog.Msg("pilot sysmsg.deletion.consumer json.Unmarshal err").Err(err).Errorx(ctx)
					continue
				}

				msgCtx := xkafka.ContextFromKafkaHeaders(kMsg.Headers)
				if ev.Uid != 0 && ev.MsgId != "" {
					err = bizz.SysNotifyBiz.DeleteSysMsg(msgCtx, ev.Uid, ev.MsgId)
					if err != nil {
						xlog.Msg("pilot sysmsg.deletion.consumer biz delete sysmsg err").Err(err).Errorx(ctx)
					}
				}
			}

			return nil
		},
	})
}
