package idgen

import (
	"context"
	"time"

	"github.com/ryanreadbooks/whimer/misc/concurrent"
	"github.com/ryanreadbooks/whimer/misc/xlog"

	foliumsdk "github.com/ryanreadbooks/folium/sdk"
)

func GetIdgen(addr string, recovery func(newIdgen foliumsdk.IClient)) foliumsdk.IClient {
	var (
		err    error
		client foliumsdk.IClient
	)

	client, err = foliumsdk.New(foliumsdk.WithGrpcOpt(addr), foliumsdk.WithDowngrade())
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
				newIdgen, err := foliumsdk.New(foliumsdk.WithGrpcOpt(addr), foliumsdk.WithDowngrade())
				if err != nil {
					xlog.Msg("re-talking to folium server failed in backgroud").Err(err).Error()
				} else {
					// replace current idgen ignoring cocurrent read-write of idgen
					recovery(newIdgen)
					return
				}
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err = client.Ping(ctx)
	if err != nil {
		xlog.Msg("can not ping idgen(folium)").Err(err).Error()
	}

	return client
}
