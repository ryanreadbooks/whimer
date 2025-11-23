package config

import (
	"github.com/ryanreadbooks/whimer/misc/xconf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

// 全局配置对象
var Conf Config

type Config struct {
	Grpc zrpc.RpcServerConf `json:"grpc"`
	Log  logx.LogConf       `json:"log"`

	MySql struct {
		User   string `json:"user"`
		Pass   string `json:"pass"`
		Addr   string `json:"addr"`
		DbName string `json:"db_name"`
	} `json:"mysql"`

	External struct {
		Grpc struct {
			Passport xconf.Discovery `json:"passport"`
			Note     xconf.Discovery `json:"note"`
			Counter  xconf.Discovery `json:"counter"`
		} `json:"grpc"`
	} `json:"external"`

	Seqer Seqer           `json:"seqer"`
	Redis redis.RedisConf `json:"redis"`

	Cron struct {
	} `json:"cron"`
}

func (c *Config) Init() error {
	return nil
}

type Seqer struct {
	Addr string `json:"addr"`
}
