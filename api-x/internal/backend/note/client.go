package note

import (
	"sync/atomic"

	"github.com/ryanreadbooks/whimer/api-x/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xgrpc/interceptor"
	notesdk "github.com/ryanreadbooks/whimer/note/sdk/v1"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

var (
	// 笔记服务
	noteAdmin    notesdk.NoteAdminServiceClient
	noteFeed     notesdk.NoteFeedServiceClient
	noteInteract notesdk.NoteInteractServiceClient

	// 是否可用
	available atomic.Bool
)

func Init(c *config.Config) {
	noteCli, err := zrpc.NewClient(
		c.Backend.Note.AsZrpcClientConf(),
		zrpc.WithUnaryClientInterceptor(interceptor.UnaryClientMetadataInject))
	if err != nil {
		logx.Errorf("external init: can not init note")
	} else {
		noteAdmin = notesdk.NewNoteAdminServiceClient(noteCli.Conn())
		noteFeed = notesdk.NewNoteFeedServiceClient(noteCli.Conn())
		noteInteract = notesdk.NewNoteInteractServiceClient(noteCli.Conn())
		available.Store(true)
	}

	initModel(c)
}

func NoteAdminServer() notesdk.NoteAdminServiceClient {
	return noteAdmin
}

func NoteInteractServer() notesdk.NoteInteractServiceClient {
	return noteInteract
}

func NoteFeedServer() notesdk.NoteFeedServiceClient {
	return noteFeed
}

func Available() bool {
	return available.Load()
}
