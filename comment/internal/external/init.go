package external

import (
	"github.com/ryanreadbooks/whimer/comment/internal/config"
	"github.com/ryanreadbooks/whimer/misc/xhttp/middleware/auth"
	notesdk "github.com/ryanreadbooks/whimer/note/sdk"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
)

var (
	// 身份认证
	auther *auth.Auth
	// 笔记服务
	noter notesdk.Note
)

func Init(c *config.Config) {
	auther = auth.MustAuther(c.External.Grpc.Passport)

	noteCli, err := zrpc.NewClient(c.External.Grpc.Note.AsZrpcClientConf())
	if err != nil {
		logx.Errorf("external init: can not init note")
	} else {
		noter = notesdk.NewNote(noteCli)
	}
}

func GetAuther() *auth.Auth {
	return auther
}

func GetNoter() notesdk.Note {
	return noter
}
