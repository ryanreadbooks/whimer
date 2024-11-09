package dep

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/passport/internal/config"

	foliumsdk "github.com/ryanreadbooks/folium/sdk"
)

var (
	// id生成器
	idgen *foliumsdk.Client
	// TODO 短信服务商接入
	sms ISmsSender

	err error
)

func Init(c *config.Config) {
	initIdGen(c)
	initSmsSender()
}

func IdGen() *foliumsdk.Client {
	return idgen
}

func SmsSender() ISmsSender {
	return sms
}

func initIdGen(c *config.Config) {
	var err error
	idgen, err = foliumsdk.NewClient(foliumsdk.WithGrpc(c.Idgen.Addr))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = idgen.Ping(ctx)
	if err != nil {
		xlog.Msg("new passport svc, can not ping idgen(folium)").Err(err).Error()
	}
}

// TODO 短信服务商
func initSmsSender() {
	sms = &logSmsSender{}
}
