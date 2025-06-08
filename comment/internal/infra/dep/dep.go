package dep

import (
	"github.com/ryanreadbooks/whimer/comment/internal/config"
	counterv1 "github.com/ryanreadbooks/whimer/counter/api/v1"
	"github.com/ryanreadbooks/whimer/misc/idgen"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	notev1 "github.com/ryanreadbooks/whimer/note/api/v1"

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
		},
	)

	counter = xgrpc.NewRecoverableClient(c.External.Grpc.Counter,
		counterv1.NewCounterServiceClient, func(nc counterv1.CounterServiceClient) {
			counter = nc
		},
	)

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
	replyIdgen = idgen.GetIdgen(c.Seqer.Addr, func(newIdgen foliumsdk.IClient) {
		replyIdgen = newIdgen
	})
}
