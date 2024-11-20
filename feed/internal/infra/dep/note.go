package dep

import (
	"github.com/ryanreadbooks/whimer/feed/internal/config"
	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	notesdk "github.com/ryanreadbooks/whimer/note/sdk/v1"
	"google.golang.org/grpc"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	// 笔记服务
	noteFeed     notesdk.NoteFeedServiceClient
	noteInteract notesdk.NoteInteractServiceClient
)

func initNote(c *config.Config) {
	var cc grpc.ClientConnInterface
	cc, err := xgrpc.NewClientConn(c.Backend.Note)
	if err != nil {
		logx.Errorf("external init: can not init note conn")
		// 后台尝试重连并重建cc
		concurrent.SafeGo(retryNoteConn(c))
		cc = xgrpc.NewUnreadyClientConn()
	}

	noteFeed = notesdk.NewNoteFeedServiceClient(cc)
	noteInteract = notesdk.NewNoteInteractServiceClient(cc)
}

func retryNoteConn(c *config.Config) func() {
	return func() {
		xgrpc.RetryConnectConn(c.Backend.Note, func(cc grpc.ClientConnInterface) {
			// we ignore concurrent read-write here
			noteFeed = notesdk.NewNoteFeedServiceClient(cc)
			noteInteract = notesdk.NewNoteInteractServiceClient(cc)
		})

	}
}

func NoteInteractServer() notesdk.NoteInteractServiceClient {
	return noteInteract
}

func NoteFeedServer() notesdk.NoteFeedServiceClient {
	return noteFeed
}
