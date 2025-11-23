package config

import (
	"github.com/ryanreadbooks/whimer/misc/obfuscate"
	"github.com/ryanreadbooks/whimer/misc/xconf"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

// 全局变量
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

	Redis redis.RedisConf `json:"redis"`

	External struct {
		Grpc struct {
			Passport xconf.Discovery `json:"passport"`
			Counter  xconf.Discovery `json:"counter"`
			Comment  xconf.Discovery `json:"comment"`
			Search   xconf.Discovery `json:"search"`
		} `json:"grpc"`
	} `json:"external"`

	Salt string `json:"salt"`

	Obfuscate struct {
		Note obfuscate.Config `json:"note"`
		Tag  obfuscate.Config `json:"tag"`
	} `json:"obfuscate"`
}

func (c *Config) Init() error {
	return nil
}
