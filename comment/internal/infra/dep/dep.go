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
	noter              notev1.NoteCreatorServiceClient // 笔记服务
	counter            counterv1.CounterServiceClient  // 计数服务
	commentIdGenerator foliumsdk.IClient
	err                error
)

func Init(c *config.Config) {
	noter = notev1.NewNoteCreatorServiceClient(
		xgrpc.NewRecoverableClientConn(c.External.Grpc.Note),
	)

	counter = counterv1.NewCounterServiceClient(
		xgrpc.NewRecoverableClientConn(c.External.Grpc.Counter),
	)

	initCommentIdgen(c)
}

func GetNoter() notev1.NoteCreatorServiceClient {
	return noter
}

func GetCounter() counterv1.CounterServiceClient {
	return counter
}

func CommentIdgen() foliumsdk.IClient {
	return commentIdGenerator
}

func initCommentIdgen(c *config.Config) {
	commentIdGenerator = idgen.GetIdgen(c.Seqer.Addr, func(newIdgen foliumsdk.IClient) {
		commentIdGenerator = newIdgen
	})
}
