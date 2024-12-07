package dep

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/comment/internal/config"
	counterv1 "github.com/ryanreadbooks/whimer/counter/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	notev1 "github.com/ryanreadbooks/whimer/note/sdk/v1"

	foliumsdk "github.com/ryanreadbooks/folium/sdk"
)

var (
	noter      notev1.NoteCreatorServiceClient // 笔记服务
	counter    counterv1.CounterServiceClient  // 计数服务
	replyIdgen foliumsdk.IClient
	err        error
)

func Init(c *config.Config) {
	noter = xgrpc.NewRecoverableClient(c.External.Grpc.Note,
		notev1.NewNoteCreatorServiceClient, func(nc notev1.NoteCreatorServiceClient) {
			noter = nc
		})

	counter = xgrpc.NewRecoverableClient(c.External.Grpc.Counter,
		counterv1.NewCounterServiceClient, func(nc counterv1.CounterServiceClient) {
			counter = nc
		})

	initReplyIdgen(c)
}

func GetNoter() notev1.NoteCreatorServiceClient {
	return noter
}

func GetCounter() counterv1.CounterServiceClient {
	return counter
}

func ReplyIdgen() foliumsdk.IClient {
	return replyIdgen
}

func initReplyIdgen(c *config.Config) {
	var err error
	replyIdgen, err = foliumsdk.New(foliumsdk.WithGrpcOpt(c.Seqer.Addr), foliumsdk.WithDowngrade())
	if err != nil {
		xlog.Msg("can not talk to folium server right now").Err(err).Error()
	}

	// try to recover idgen in background
	concurrent.SafeGo(func() {
		xlog.Msg("re-talking to folium server started in background")
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				newIdgen, err := foliumsdk.New(foliumsdk.WithGrpcOpt(c.Seqer.Addr), foliumsdk.WithDowngrade())
				if err != nil {
					xlog.Msg("re-talking to folium server failed in backgroud").Err(err).Error()
				} else {
					// replace current idgen ignoring cocurrent read-write of idgen
					replyIdgen = newIdgen
					return
				}
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = replyIdgen.Ping(ctx)
	if err != nil {
		xlog.Msg("new passport svc, can not ping idgen(folium)").Err(err).Error()
	}
}
