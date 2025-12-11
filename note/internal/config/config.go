package config

import (
	"github.com/ryanreadbooks/whimer/misc/obfuscate"
	"github.com/ryanreadbooks/whimer/misc/xconf"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

// 全局变量
var Conf Config

type Config struct {
	Http rest.RestConf      `json:"http"`
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
			Passport  xconf.Discovery `json:"passport"`
			Counter   xconf.Discovery `json:"counter"`
			Comment   xconf.Discovery `json:"comment"`
			Search    xconf.Discovery `json:"search"`
			Conductor xconf.Discovery `json:"conductor"`
		} `json:"grpc"`
	} `json:"external"`

	Salt string `json:"salt"`

	Obfuscate struct {
		Note obfuscate.Config `json:"note"`
		Tag  obfuscate.Config `json:"tag"`
	} `json:"obfuscate"`

	DevCallbacks DevCallbacks `json:"dev_callbacks"`
}

func (c *Config) Init() error {
	return nil
}

type DevCallbacks struct {
	NoteProcessCallback string `json:"note_process_callback"`
}
