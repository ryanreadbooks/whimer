package dep

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xlog"
	"github.com/ryanreadbooks/whimer/passport/internal/config"

	foliumsdk "github.com/ryanreadbooks/folium/sdk"
)

var (
	idgen foliumsdk.IClient // id生成器
	sms   ISmsSender        // TODO 短信服务商接入
	err   error
)

func Init(c *config.Config) {
	initIdGen(c)
	initSmsSender()
}

func IdGen() foliumsdk.IClient {
	return idgen
}

func SmsSender() ISmsSender {
	return sms
}

func initIdGen(c *config.Config) {
	var err error
	idgen, err = foliumsdk.New(foliumsdk.WithGrpcOpt(c.Idgen.Addr), foliumsdk.WithDowngrade())
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
				newIdgen, err := foliumsdk.New(foliumsdk.WithGrpcOpt(c.Idgen.Addr), foliumsdk.WithDowngrade())
				if err != nil {
					xlog.Msg("re-talking to folium server failed in backgroud").Err(err).Error()
				} else {
					// replace current idgen ignoring cocurrent read-write of idgen
					idgen = newIdgen
				}
			}
		}
	})

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
