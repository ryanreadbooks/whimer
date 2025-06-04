package config

import (
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/zeromicro/go-zero/rest"
)

type Websocket struct {
	ReadTimeout    xtime.SDuration `json:"read_timeout"`
	WriteTimeout   xtime.SDuration `json:"write_timeout"`
	BusyThreshold  int             `json:"busy_threshold"` // 判定为忙碌的连接占比阈值
	MaxConnAllowed int             `json:"max_conn_allowed"`
}

type Config struct {
	Http     rest.RestConf `json:"http"`
	WsServer *Websocket    `json:"ws_server"`
}
