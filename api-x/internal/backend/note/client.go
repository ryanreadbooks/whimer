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
	noter notesdk.NoteServiceClient
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
		noter = notesdk.NewNoteServiceClient(noteCli.Conn())
		available.Store(true)
	}

	initModel(c)
}

func GetNoter() notesdk.NoteServiceClient {
	return noter
}

func Available() bool {
	return available.Load()
}
