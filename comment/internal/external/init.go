package external

import (
	"github.com/ryanreadbooks/whimer/comment/internal/config"
	countersdk "github.com/ryanreadbooks/whimer/counter/sdk/v1"
	"github.com/ryanreadbooks/whimer/misc/xgrpc"
	notesdk "github.com/ryanreadbooks/whimer/note/sdk/v1"
	"github.com/ryanreadbooks/whimer/passport/sdk/middleware/auth"

	"github.com/zeromicro/go-zero/core/logx"
)

var (
	// 身份认证
	auther *auth.Auth
	// 笔记服务
	noter notesdk.NoteAdminServiceClient
	// 计数服务
	counter countersdk.CounterServiceClient
)

func Init(c *config.Config) {
	auther = auth.MustAuther(c.External.Grpc.Passport)

	var err error

	noter, err = xgrpc.NewClient(c.External.Grpc.Note,
		notesdk.NewNoteAdminServiceClient)
	if err != nil {
		logx.Errorf("external init: can not init noter: %v", err)
	}

	counter, err = xgrpc.NewClient(c.External.Grpc.Counter,
		countersdk.NewCounterServiceClient)
	if err != nil {
		logx.Errorf("external init: can not init counter: %v", err)
	}
}

func GetAuther() *auth.Auth {
	return auther
}

func GetNoter() notesdk.NoteAdminServiceClient {
	return noter
}

func GetCounter() countersdk.CounterServiceClient {
	return counter
}
