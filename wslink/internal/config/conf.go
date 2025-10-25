package config

import (
	"github.com/ryanreadbooks/whimer/misc/xconf"
	"github.com/ryanreadbooks/whimer/misc/xtime"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// 全局配置对象
var Conf Config

type Websocket struct {
	ReadTimeout    xtime.SDuration `json:"read_timeout"`
	WriteTimeout   xtime.SDuration `json:"write_timeout"`
	BusyThreshold  int             `json:"busy_threshold"` // 判定为忙碌的连接占比阈值
	MaxConnAllowed int             `json:"max_conn_allowed"`
}

type Config struct {
	System struct {
		Shutdown struct {
			WaitTime int `json:"wait_time"` // sec
		} `json:"shutdown"`
		ConnShard int `json:"conn_shard"`
	} `json:"system"`

	Http     rest.RestConf      `json:"http"`
	Grpc     zrpc.RpcServerConf `json:"grpc"`
	Log      logx.LogConf       `json:"log"`
	WsServer *Websocket         `json:"ws_server"`

	Redis redis.RedisConf

	Backend struct {
		Passport xconf.Discovery `json:"passport"`
	} `json:"backend"`
}
